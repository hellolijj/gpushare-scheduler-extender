package policy

import (
	"encoding/json"
	"fmt"

	gputype "github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/types"
	"os"
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/utils"
	"log"
	"sort"
)

type staticRunner struct{
	configPath string
}

// NewSimplePolicy creates a new SimplePolicy.
func NewStaticRunner(path string) Run {
	if len(path) == 0 {
		return nil
	}
	// TODO add more checks
	return &staticRunner{configPath:path}
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
	
	nodeType := utils.GetNodeTypeFromAnnotation(n.GetNode())
	if len(nodeType) == 0 {
		log.Printf("warn: can not get node type")
	}
	
	// 构造可用 devices
	devices := []int{}
	for _, dev := range n.GetDevs() {
		if dev.IsUsed() == false {
			devices = append(devices, dev.GetDevId())
		}
	}
	
	nodeConfig, err := loadNodeTypeConfig(s.configPath)
	log.Printf("debug: get nodeconfig: %v", nodeConfig)
	if err != nil {
		return nil, 0, err
	}
	
	validSet, ok := nodeConfig[nodeType]
	if !ok {
		log.Printf("warn: no avaliable gpu config %v for node type %s", validSet, nodeType)
	}
	log.Printf("debug: get valid set: %v", validSet)
	
	// 1. 找到与配置文件相同策略20分
	validGpuSets, ok := validSet[req]
	if ok {
		for _, validGpuSet := range validGpuSets {
			if isChinldSet(validGpuSet, devices) {
				log.Printf("info: select a 20 score policy %v", validGpuSet)
				return validGpuSet, 20, nil
			}
		}
	}
	
	// 2. 以req=3为例， req++, 如果找到 req+ 的配置策略 并且可分配，15分
	virtualReq := req
	for {
		virtualReq++
		validGpuSets, ok := validSet[virtualReq]
		if ok {
			for _, validGpuSet := range validGpuSets {
				if isChinldSet(validGpuSet, devices) {
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
func isChinldSet(child, parent []int) bool {
	sort.Ints(child)
	sort.Ints(parent)
	if len(child) > len(parent) {
		return false
	}
	for i := 0; i < len(child); i++ {
		if child[i] != parent[i] {
			return false
		}
	}
	return true
}