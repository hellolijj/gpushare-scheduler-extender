package types

import (
	"log"
	"strings"
	"sync"

	"encoding/json"
	"strconv"

	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
	"k8s.io/api/core/v1"
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/utils"
)

const (
	OptimisticLockErrorMsg = "the object has been modified; please apply your changes to the latest version and try again"
)

// NodeInfo is node level aggregateyd information.
type NodeInfo struct {
	name        string
	node        *v1.Node
	devs        map[int]*DeviceInfo
	gpuUsed     int
	gpuCount    int
	gpuTopology [][]TopologyType
	rwmu        *sync.RWMutex
}

type TopologyType nvml.P2PLinkType

func (t TopologyType) Desc() string {
	return nvml.P2PLinkType(t).String()
}

func (t TopologyType) Abbr() string {
	return utils.GetGPUAbbr(nvml.P2PLinkType(t))
}

// Create Node Level
func NewNodeInfo(node *v1.Node) *NodeInfo {
	log.Printf("debug: NewNodeInfo() creates nodeInfo for %s", node.Name)

	devMap := map[int]*DeviceInfo{}
	for i := 0; i < utils.GetGPUCountInNode(node); i++ {
		devMap[i] = newDeviceInfo(i)
	}

	gpuTopology := getGPUTopologyFromNode(node, devMap)

	nodeInfo := &NodeInfo{
		name:        node.Name,
		node:        node,
		devs:        devMap,
		gpuCount:    utils.GetGPUCountInNode(node),
		gpuTopology: gpuTopology,

		rwmu: new(sync.RWMutex),
	}

	log.Printf("debug: node %s has nodeinfo %v", node.Name, nodeInfo)
	log.Printf("debug: node %s has topology %v", node.Name, gpuTopology)
	return nodeInfo
}

// 从node annotaion 里获取gpu topology
func getGPUTopologyFromNode(node *v1.Node, devs map[int]*DeviceInfo) [][]TopologyType {
	// init gpuTopology
	topology := make([][]TopologyType, len(devs))

	if !utils.IsGPUTopologyNode(node) {
		return topology
	}

	for i := 0; i < len(devs); i++ {
		topology[i] = make([]TopologyType, len(devs))
	}

	log.Printf("debug: node %s has annotation %v", node.Name, node.Annotations)

	gpuTopology, ok := node.Annotations[utils.EnvGPUAnnotation]
	if !ok {
		return topology
	}

	gpuTopologyMap := map[string]string{}
	json.Unmarshal([]byte(gpuTopology), &gpuTopologyMap)
	if len(gpuTopologyMap) == 0 {
		return topology
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
		topology[gpu1][gpu2] = TopologyType(utils.GetGPULinkFromDescAndAbbr(topoAbbr))
		topology[gpu2][gpu1] = TopologyType(utils.GetGPULinkFromDescAndAbbr(topoAbbr))
	}

	return topology
}

func (n *NodeInfo) GetName() string {
	return n.name
}

func (n *NodeInfo) GetDevs() map[int]*DeviceInfo {
	return n.devs
}

func (n *NodeInfo) GetNode() *v1.Node {
	return n.node
}

func (n *NodeInfo) GetGPUCount() int {
	return n.gpuCount
}

func (n *NodeInfo) GetGPUTopology() [][]TopologyType {
	return n.gpuTopology
}

func (n *NodeInfo) GetGPUUsedCount() int {
	count := 0
	for _, dev := range n.devs {
		if dev.isUsed {
			count++
		}
	}
	return count
}

func (n *NodeInfo) RemovePod(pod *v1.Pod) {
	n.rwmu.Lock()
	defer n.rwmu.Unlock()

	ids := utils.GetGPUIDFromAnnotation(pod)
	log.Printf("warn: Pod remove ids %v", ids)
	for _, id := range ids {
		if id >= 0 {
			dev, found := n.devs[id]
			if !found {
				log.Printf("warn: Pod %s in ns %s failed to find the GPU ID %d in node %s", pod.Name, pod.Namespace, id, n.name)
			} else {
				dev.RemovePod(pod)
			}
		} else {
			log.Printf("warn: Pod %s in ns %s is not set the GPU ID %d in node %s", pod.Name, pod.Namespace, id, n.name)
		}
	}
}

// Add the Pod which has the GPU id to the node
func (n *NodeInfo) AddOrUpdatePod(pod *v1.Pod) (added bool) {
	n.rwmu.Lock()
	defer n.rwmu.Unlock()

	ids := utils.GetGPUIDFromAnnotation(pod)
	log.Printf("debug: addOrUpdatePod() Pod %s in ns %s with the GPU IDs %v should be added to device map",
		pod.Name,
		pod.Namespace,
		ids)
	if len(ids) >= 0 {
		for _, id := range ids {
			dev, found := n.devs[id]
			if !found {
				log.Printf("warn: Pod %s in ns %s failed to find the GPU ID %d in node %s", pod.Name, pod.Namespace, id, n.name)
			} else {
				dev.AddPod(pod)
				added = true
			}
		}
	} else {
		log.Printf("warn: Pod %s in ns %s is not set the GPU ID in node %s", pod.Name, pod.Namespace, n.name)
	}
	return added
}

// check if the pod can be allocated on the node
func (n *NodeInfo) Assume(pod *v1.Pod) (allocatable bool) {
	allocatable = false

	n.rwmu.RLock()
	defer n.rwmu.RUnlock()

	availableGPUs := n.GetAvailableGPUs()
	reqGPU := utils.GetGPUCountFromPodResource(pod)
	log.Printf("debug: AvailableGPUs: %v in node %s", availableGPUs, n.name)
	log.Printf("debug: requestGPUs: %v in node %s", reqGPU, n.name)

	if availableGPUs > 0 && availableGPUs-reqGPU >= 0 {
		allocatable = true
	}

	return
}

func (n *NodeInfo) GetAvailableGPUs() (availableGPUs int) {
	allGPUs := n.GetAllGPUs()
	usedGPUs := n.GetGPUUsedCount()
	availableGPUs = allGPUs - usedGPUs
	return availableGPUs
}

func (n *NodeInfo) GtUsedGPUs() (usedGPUs int) {
	usedGPUs = n.GetGPUUsedCount()
	log.Printf("debug: getUsedGPUs: %v in node %s, and devs %v", usedGPUs, n.name, n.devs)
	return usedGPUs
}

func (n *NodeInfo) GetAllGPUs() (allGPUs int) {
	allGPUs = n.GetGPUCount()
	log.Printf("debug: getAllGPUs: %v in node %s, and dev %v", allGPUs, n.name, n.devs)
	return allGPUs
}

var (
	nodeype = map[string]nvml.P2PLinkType{
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
)
