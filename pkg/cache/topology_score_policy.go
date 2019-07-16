package cache

import (
	"fmt"
	"log"

	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
)

type topologyPolicy struct{}

// NewSimplePolicy creates a new SimplePolicy.
func NewTopologyPolicy() Policy {
	return &topologyPolicy{}
}

// Allocate GPUs following a simple policy.
func (p *topologyPolicy) Score(n *NodeInfo, ques int) int {
	availableGPUs := n.getAvailableGPUs()

	if ques <= 0 || ques > availableGPUs {
		return 0
	}
	return 0

}

// PreAllocate 计算分配方案，及该方案的打分
func (p *topologyPolicy) PreAllocate(n *NodeInfo, req int) (ids []int, score int, err error) {
	availableGPUs := n.getAvailableGPUs()
	if req <= 0 || req > availableGPUs {
		err = fmt.Errorf("rqu gpu count %v is invalidl", req)
		return nil, 0, err
	}

	if req == 1 {
		minScoreId, score, err := getMinScoreGpu(n)
		if err != nil {
			return nil, 0, err
		}
		ids = append(ids, minScoreId)
		// TODO: score 越小越好
		score = 1000 - score
		return nil, 0, err
	}
	
	log.Printf("request gpus counts %v is more than 1, ", req)
	ids, score, err = getMaxScoreLink(n)
	if req == 2 {
		return
	}
	
	// 接下来是图中寻找最小生成树问题，参见 prim 算法
	
	return
	
}


// 获取 gpu topo 中 离拓扑中心最远到卡
func getMinScoreGpu(n *NodeInfo) (id int, score int, err error) {
	minScore := 10000
	minDevIdx := -1
	for _, dev1 := range n.devs {
		if dev1.isUsed == false {
			score := 0
			for _, dev2 := range n.devs {
				if dev1 != dev2 && dev2.isUsed == false {
					score += calculateGPUPairScore(n, dev1.idx, dev2.idx)
				}
			}
			if score < minScore {
				minScore = score
				minDevIdx = dev1.idx
			}
		}
	}
	if minDevIdx == -1 {
		err = fmt.Errorf("the node %s can't choose unused gpu", n.node.Name)
		return -1, 0, nil
	}
	
	return minDevIdx, minScore, nil
}


// 计算两块 gpu 之间的分数
func calculateGPUPairScore(n *NodeInfo, gpu0 int, gpu1 int) int {
	if gpu0 < 0 || gpu0 > len(n.devs)-1 || gpu1 < 0 || gpu1 > len(n.devs)-1 {
		return 0
	}

	if gpu0 == gpu1 {
		log.Printf("can't calculate the same gpu score: %v", gpu1)
		return 0
	}

	if n.devs[gpu0].isUsed || n.devs[gpu1].isUsed {
		log.Printf("gpu can't be used: gpu%d %v, gpu%d, %v", gpu0, n.devs[gpu0].isUsed, gpu1, n.devs[gpu1].isUsed)
		return 0
	}

	score := 0
	topology := n.GetGPUTopology()
	linkType := topology[gpu0][gpu1]

	switch nvml.P2PLinkType(linkType) {
	case nvml.P2PLinkCrossCPU:
		score += 10
	case nvml.P2PLinkSameCPU:
		score += 20
	case nvml.P2PLinkHostBridge:
		score += 30
	case nvml.P2PLinkMultiSwitch:
		score += 40
	case nvml.P2PLinkSingleSwitch:
		score += 50
	case nvml.P2PLinkSameBoard:
		score += 60
	case nvml.SingleNVLINKLink:
		score += 100
	case nvml.TwoNVLINKLinks:
		score += 200
	case nvml.ThreeNVLINKLinks:
		score += 300
	case nvml.FourNVLINKLinks:
		score += 400
	case nvml.FiveNVLINKLinks:
		score += 500
	case nvml.SixNVLINKLinks:
		score += 600
	}

	return score
}

// 计算分数最高的边，返回两个节点
func getMaxScoreLink(n *NodeInfo) (ids []int, score int, err error) {
	var maxDevIdx1, maxDevIdx2 int
	maxScore := -1
	
	for _, dev1 := range n.devs {
		if dev1.isUsed == false {
			for _, dev2 := range n.devs {
				if dev1 != dev2 && dev2.isUsed == false {
					score := calculateGPUPairScore(n, dev1.idx, dev2.idx)
					if score > maxScore {
						maxScore = score
						maxDevIdx1, maxDevIdx2 = dev1.idx, dev2.idx
					}
				}
			}
		}
	}
	
	if maxScore == -1 {
		err = fmt.Errorf("the node %s can't choose unused gpu", n.node.Name)
		return []int{}, 0, nil
	}
	
	return []int{maxDevIdx1, maxDevIdx2}, maxScore, nil
}