package policy

import (
	"fmt"
	gputype "github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/types"

)

type staticRunner struct{}

// NewSimplePolicy creates a new SimplePolicy.
func NewStaticRunner() Run {
	return &staticRunner{}
}

func (s *staticRunner) Score(n *gputype.NodeInfo, req int) (int, error) {
	ids, _, err := s.PreAllocate(n, req)
	if err != nil || len(ids) == 0{
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

	validSets := ValidSets()
	
	devices := []int{}
	for _, dev := range n.GetDevs() {
		if dev.IsUsed() == false {
			devices = append(devices, dev.GetDevId())
		}
	}

	res := findGPUSet(devices, req, validSets["shenglong"][req])
	if len(res) > 0 {
		return res, 10, nil
	}

	return []int{}, 0, fmt.Errorf("no is invalid gpu")
}

// Find a GPU set of size 'size' in the list of devices that is contained in 'validSets'.
func findGPUSet(devices []int, size int, validSets [][]int) []int {
	solutionSet := []int{}

	for _, validSet := range validSets {
		for _, i := range validSet {
			for _, device := range devices {
				if device == i {
					solutionSet = append(solutionSet, device)
					break
				}
			}
		}

		if len(solutionSet) == size {
			break
		}

		solutionSet = []int{}
	}

	return solutionSet
}

func NodeTypeConfig() map[string]map[int][][]int {
	
	validSets := make(map[string]map[int][][]int)
	
	shenglongValidConfig := map[int][][]int{
		1: {{0}, {1}, {2}, {3}, {4}, {5}, {6}, {7}},
		2: {{0, 2}, {1, 3}, {4, 6}, {5, 7}},
		4: {{0, 1, 2, 3}, {4, 5, 6, 7}},
		8: {{0, 1, 2, 3, 4, 5, 6, 7}},
	}
	
	validSets["shenglong"] = shenglongValidConfig
	
	return validSets
}

// 从配置文件中加载 node type 配置
func loadNodeTypeConfig() {

}