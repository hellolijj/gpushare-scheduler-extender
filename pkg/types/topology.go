package types

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/utils"
	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
	"k8s.io/api/core/v1"
)

var (
	abbrGpu = map[string]nvml.P2PLinkType{
		"PSB":  1,
		"PIX":  2,
		"PXB":  3,
		"PHB":  4,
		"NODE": 5,
		"SYS":  6,
		"NV1":  7,
		"NV2":  8,
		"NV3":  9,
		"NV4":  10,
		"NV5":  11,
		"NV6":  12,
	}
	gpuAbbr = map[nvml.P2PLinkType]string{
		1:  "PSB",
		2:  "PIX",
		3:  "PXB",
		4:  "PHB",
		5:  "NODE",
		6:  "SYS",
		7:  "NV1",
		8:  "NV2",
		9:  "NV3",
		10: "NV4",
		11: "NV5",
		12: "NV6",
	}
)

type TopologyType nvml.P2PLinkType

func (t *TopologyType) Desc() string {
	return nvml.P2PLinkType(*t).String()
}

func (t *TopologyType) Abbr() string {
	abbr, ok := gpuAbbr[nvml.P2PLinkType(*t)]
	if !ok {
		return ""
	}
	return abbr
}

type Topology [][]TopologyType

// 从node annotaion 里获取gpu topology
func NewTopology(node *v1.Node) Topology {
	topology := getOriginGpuTopology(node)
	return topology
}


// 存在一个任务使用到 gpu 卡 ids
// TODO: ADD LOCK HERE
func (t Topology) ConsumeId(id int) {
	numGpus := len(t)
	if id < 0 || id >= numGpus{
		log.Printf("warn: invaild consume id %v", id)
		return
	}
	for i := 0; i < numGpus; i++ {
		if i == id{
			for k := 0; k < numGpus; k++ {
				t[i][k] = 0
			}
			continue
		}
		for j := 0; j < numGpus; j ++ {
			if j == id {
				t[i][j] = 0
			}
		}
	}
}

// TODO: ADD LOCK HERE
func (t Topology) RecoveryId(id int, newTopology Topology) {
	numGpus := len(t)
	if id < 0 || id >= numGpus{
		log.Printf("warn: invaild consume id %v", id)
		return
	}
	if len(newTopology) != numGpus {
		log.Printf("warn: invaild new topology %v", newTopology)
		return
	}
	
	for i := 0; i < numGpus; i++ {
		if i == id{
			for k := 0; k < numGpus; k++ {
				t[i][k] = newTopology[i][k]
			}
			continue
		}
		for j := 0; j < numGpus; j ++ {
			if j == id {
				t[i][j] = newTopology[i][j]
			}
		}
	}
}


func getGPULinkFromAbbr(abbr string) nvml.P2PLinkType {
	abbrKey, ok := abbrGpu[abbr]
	if !ok {
		return 0
	}
	return abbrKey
}

func getOriginGpuTopology(node *v1.Node) Topology{
	if !utils.IsGPUTopologyNode(node) {
		return nil
	}
	
	numGpus := utils.GetGPUCountInNode(node)
	if numGpus == 0 {
		return nil
	}
	
	topology := make(Topology, numGpus)
	for i := 0; i < numGpus; i++ {
		topology[i] = make([]TopologyType, numGpus)
	}
	
	gpuTopologyJson, ok := node.Annotations[utils.EnvGPUAnnotation]
	if !ok {
		return nil
	}
	
	gpuTopologyMap := map[string]string{}
	json.Unmarshal([]byte(gpuTopologyJson), &gpuTopologyMap)
	if len(gpuTopologyMap) == 0 {
		return nil
	}
	
	for k, v := range gpuTopologyMap {
		var gpu1, gpu2 int
		var topoAbbr string
		
		// 使用 _ 分割获取结果。
		gpuSplitRes := strings.Split(k, "_")
		
		if len(gpuSplitRes) != 4 {
			log.Printf("warn: annotation topology split error")
			continue
		}
		
		topoAbbr = gpuSplitRes[1]
		gpu1, err := strconv.Atoi(gpuSplitRes[2])
		if err != nil {
			log.Printf("warn: get gpu1 error: %v", err)
			continue
		}
		gpu2, err = strconv.Atoi(gpuSplitRes[3])
		if err != nil {
			log.Printf("warn: get gpu2 error: %v", err)
			continue
		}
		
		log.Printf("debug: from annotaion %s to get gputoplogy gpu%d and gpu%d 's relations is desc: %s abbr: %s", k, gpu1, gpu2, v, topoAbbr)
		topology[gpu1][gpu2] = TopologyType(getGPULinkFromAbbr(topoAbbr))
		topology[gpu2][gpu1] = TopologyType(getGPULinkFromAbbr(topoAbbr))
	}
	
	return topology
}