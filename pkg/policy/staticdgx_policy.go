package policy

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"

	gputype "github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/types"
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/utils"

	"k8s.io/api/core/v1"
)

type staticRunner struct {
	configPath string
}

// NewSimplePolicy creates a new SimplePolicy.
func NewStaticRunner(path string) Run {
	if len(path) == 0 {
		return nil
	}
	// TODO add more checks
	return &staticRunner{configPath: path}
}

func (s *staticRunner) Score(n *gputype.NodeInfo, req int) (int, error) {
	ids, _, err := s.PreAllocate(n, req)
	if err != nil || len(ids) == 0 {
		return 0, err
	}

	// 如果要有返回值，统一打分10
	return 10, nil
}

func (s *staticRunner) Allocate(n *gputype.NodeInfo, req int) ([]int, error) {
	ids, _, err := s.PreAllocate(n, req)
	return ids, err
}

// PreAllocate 计算分配方案，及该方案的打分
func (s *staticRunner) PreAllocate(n *gputype.NodeInfo, req int) (ids []int, score int, err error) {
	availableGPUs := n.GetAvailableGPUs()
	if req <= 0 || req > availableGPUs {
		err = fmt.Errorf("rqu gpu count %v is invalid", req)
		return nil, 0, err
	}

	// 构造可用 devices
	devices := []int{}
	for _, dev := range n.GetDevs() {
		if dev.IsUsed() == false {
			devices = append(devices, dev.GetDevId())
		}
	}
	log.Printf("debug: get devices: %v", devices)

	validSet := s.GetStaticConfig(n.GetNode())
	log.Printf("debug: get valid set: %v", validSet)

	// 1. 找到与配置文件相同策略20分
	validGpuSets, ok := validSet[req]
	if ok {
		log.Printf("debug: get validGpuSets %v", validGpuSets)
		for _, validGpuSet := range validGpuSets {
			if isSubInSet(validGpuSet, devices) {
				log.Printf("info: select a 20 score policy %v", validGpuSet)
				return validGpuSet, 20, nil
			}
		}
	} else {
		log.Println("debug: unable to get validGpuSets")
	}

	// 2. 以req=3为例， req++, 如果找到 req+ 的配置策略 并且可分配，15分
	virtualReq := req
	for {
		virtualReq++
		validGpuSets, ok := validSet[virtualReq]
		if ok {
			for _, validGpuSet := range validGpuSets {
				if isSubInSet(validGpuSet, devices) {
					// 从 vailidGpuSet 中随机选择
					log.Printf("info: select a 15 score policy %v", validGpuSet[:req])
					return validGpuSet[:req], 15, nil
				}
			}
		}
		if virtualReq > len(devices) {
			break
		}
	}

	// 3. 随机选择
	log.Printf("info: choose a random schem %v", devices[:req])
	return devices[:req], 10, fmt.Errorf("no is invalid gpu")
}

func (s *staticRunner) GetStaticConfig(n *v1.Node) map[int][][]int {
	nodeType := utils.GetNodeTypeFromAnnotation(n)
	if len(nodeType) == 0 {
		log.Printf("warn: can not get node type")
		return nil
	}
	nodeConfig, err := loadNodeTypeConfig(s.configPath)
	log.Printf("debug: get nodeconfig: %v", nodeConfig)
	if err != nil {
		return nil
	}

	validSet, ok := nodeConfig[nodeType]
	if !ok {
		log.Printf("warn: no avaliable gpu config %v for node type %s", validSet, nodeType)
		return nil
	}
	return validSet
}

// 从配置文件中加载 node type 配置
func loadNodeTypeConfig(path string) (map[string]map[int][][]int, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("config path can not be empty")
	}
	//TODO: ADD more check

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var nodeTypeConfig map[string]map[int][][]int

	err = json.NewDecoder(f).Decode(&nodeTypeConfig)
	if err != nil {
		return nil, fmt.Errorf("config file error: %v", f.Name())
	}

	return nodeTypeConfig, nil
}

// child is shorter than parent
// whether child
func isSubInSet(child, parent []int) bool {
	sort.Ints(child)
	sort.Ints(parent)
	log.Printf("debug: chind %v parent %v", child, parent)
	if len(child) > len(parent) {
		return false
	}
	for i := 0; i < len(child); i++ {
		if !isNInSet(child[i], parent) {
			return false
		}
	}
	return true
}

// 判断 元素 n 是否存在于 set 中
func isNInSet(n int, s []int) bool {
	if len(s) == 0 {
		return false
	}
	for _, v := range s {
		if v == n {
			return true
		}
	}
	return false
}
