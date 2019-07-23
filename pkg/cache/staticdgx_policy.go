package cache

import "fmt"

// GPUType represents the valid set of GPU
// types a Static DGX policy can be created for.
type GPUType int

// Valid GPUTypes
const (
	GPUTypePascal GPUType = iota // Pascal GPUs
	GPUTypeVolta
)

// Policy Definitions
type staticDGX1PascalPolicy struct{}
type staticDGX1VoltaPolicy struct{}
type staticDGX2VoltaPolicy struct{}

// NewStaticDGX1Policy creates a new StaticDGX1Policy for gpuType.
func NewStaticDGX1Policy(gpuType GPUType) Policy {
	if gpuType == GPUTypePascal {
		return &staticDGX1PascalPolicy{}
	}
	if gpuType == GPUTypeVolta {
		return &staticDGX1VoltaPolicy{}
	}
	return nil
}

// NewStaticDGX2Policy creates a new StaticDGX2Policy.
func NewStaticDGX2Policy() Policy {
	return &staticDGX1VoltaPolicy{}
}

// PreAllocate 计算分配方案，及该方案的打分
func (s *staticDGX1PascalPolicy) PreAllocate(n *NodeInfo, req int) (ids []int, score int, err error) {
	availableGPUs := n.getAvailableGPUs()
	if req <= 0 || req > availableGPUs {
		err = fmt.Errorf("rqu gpu count %v is invalid", req)
		return nil, 0, err
	}

	validSets := map[int][][]int{
		1: {{0}, {1}, {2}, {3}, {4}, {5}, {6}, {7}},
		2: {{0, 2}, {1, 3}, {4, 6}, {5, 7}},
		4: {{0, 1, 2, 3}, {4, 5, 6, 7}},
		8: {{0, 1, 2, 3, 4, 5, 6, 7}},
	}

	devices := []int{}
	for _, dev := range n.devs {
		if dev.isUsed == false {
			devices = append(devices, dev.idx)
		}
	}

	res := findGPUSet(devices, req, validSets[req])
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
