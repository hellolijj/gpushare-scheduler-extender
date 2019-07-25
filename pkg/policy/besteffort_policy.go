package policy

import (
	"fmt"
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/utils"
)

type bestRunner struct{}

func NewBestRunner() Run {
	return &bestRunner{}
}


// Allocate GPUs following a simple policy.
func (b *bestRunner) Score(n *utils.NodeInfo, req int) (int, error) {
	_, score, err := b.PreAllocate(n, req)
	return score, err
}

func (b *bestRunner) Allocate(n *utils.NodeInfo, req int) ([]int, error) {
	ids, _, err := b.PreAllocate(n, req)
	return ids, err
}

// bestRunner 计算分配方案，及该方案的打分
func (b *bestRunner) PreAllocate(n *utils.NodeInfo, req int) (ids []int, score int, err error) {
	availableGPUs := n.GetAvailableGPUs()
	if req <= 0 || req > availableGPUs {
		err = fmt.Errorf("rqu gpu count %v is invalid", req)
		return nil, 0, err
	}
	
	devices := []int{}
	for _, dev := range n.GetDevs() {
		if dev.IsUsed() == false {
			devices = append(devices, dev.GetDevId())
		}
	}
	
	bestPartition := [][]int{}
	bestScore := 0
	iterateGPUPartitions(devices, req, func(candidate [][]int) {
		score := calculateGPUPartitionScore(n, candidate)
		if score > bestScore || bestPartition == nil {
			bestPartition = candidate
			bestScore = score
		}
	})
	
	// Find the highest scoring GPU set in the highest scoring GPU partition.
	bestSet := bestPartition[0]
	bestScore = calculateGPUSetScore(n, bestSet)
	for i := 1; i < len(bestPartition); i++ {
		score := calculateGPUSetScore(n, bestPartition[i])
		if score > bestScore {
			bestSet = bestPartition[i]
			bestScore = score
		}
	}
	
	return bestSet, bestScore, nil
}


func iterateGPUPartitions(devices []int, size int, callback func([][]int)) {
	if size <= 0 {
		return
	}
	
	if size > len(devices) {
		return
	}
	
	// Optimize for the case when size == 1.
	if size == 1 {
		for _, device := range devices {
			callback([][]int{[]int{device}})
		}
		return
	}
	
	devices = gpuSetCopyAndAddPadding(devices, size)
	padding := gpuSetCountPadding(devices)
	
	// We wrap the recursive call to make use of an 'accum' variable to
	// build out each partition as the recursion progresses.
	var iterate func(devices []int, size int, accum [][]int)
	iterate = func(devices []int, size int, accum [][]int) {
		// Padding should ensure that his never happens.
		if size > len(devices) {
			panic("Internal error in best effort allocation policy")
		}
		
		// Base case once we've reached 'size' number of devices.
		if size == len(devices) {
			callback(append(accum, devices))
			return
		}
		
		iterateGPUSets(devices[1:], size-1, func(set []int) {
			set = append([]int{devices[0]}, set...)
			
			p := gpuSetCountPadding(set)
			if !(p == 0 || p == padding) {
				return
			}
			
			remaining := []int{}
			for _, gpu := range devices {
				if !gpuSetContains(set, gpu) {
					remaining = append(remaining, gpu)
				}
			}
			
			iterate(remaining, size, append(accum, set))
		})
	}
	
	iterate(devices, size, [][]int{})
}

func gpuSetCopyAndAddPadding(gpuSet []int, size int) []int {
	if size <= 0 {
		return []int{}
	}
	
	gpus := append([]int{}, gpuSet...)
	for len(gpus)%size != 0 {
		gpus = append(gpus, -1)
	}
	
	return gpus
}

// Count the amount of padding in the GPU set.
func gpuSetCountPadding(gpuSet []int) int {
	count := 0
	
	for i := range gpuSet {
		if gpuSet[i] == -1 {
			count++
		}
	}
	
	return count
}

func iterateGPUSets(devices []int, size int, callback func([]int)) {
	if size <= 0 {
		return
	}
	
	if size > len(devices) {
		return
	}
	
	level := 0
	indices := make([]int, size)
	set := make([]int, size)
	
	for {
		if indices[level] == len(devices) {
			if level == 0 {
				break
			}
			
			level--
			indices[level]++
			continue
		}
		
		set[level] = devices[indices[level]]
		
		if level < (size - 1) {
			level++
			indices[level] = indices[level-1] + 1
			continue
		}
		
		callback(set)
		indices[level]++
	}
}

// Check to see if a specific GPU is contained in a GPU set.
func gpuSetContains(gpuSet []int, gpu int) bool {
	for i := range gpuSet {
		if gpuSet[i] == gpu {
			return true
		}
	}
	return false
}


// Get the total score of a set of GPUs. The score is calculated as the sum of
// the scores calculated for each pair of GPUs in the set.
func calculateGPUSetScore(n *utils.NodeInfo, gpuSet []int) int {
	score := 0
	
	iterateGPUSets(gpuSet, 2, func(gpus []int) {
		score += calculateGPUPairScore(n, gpus[0], gpus[1])
	})
	
	return score
}

// Get the total score of a GPU partition. The score is calculated as the sum
// of the scores calculated for each set of GPUs within the partition.
func calculateGPUPartitionScore(n *utils.NodeInfo, gpuPartition [][]int) int {
	score := 0
	
	for _, gpuSet := range gpuPartition {
		score += calculateGPUSetScore(n, gpuSet)
	}
	
	return score
}