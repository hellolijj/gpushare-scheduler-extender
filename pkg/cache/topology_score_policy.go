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
func (p *topologyPolicy) Score(n *NodeInfo, req int) (int, error) {
	availableGPUs := n.getAvailableGPUs()

	if req <= 0 || req > availableGPUs {
		err := fmt.Errorf("rqu gpu %v is invalid", req)
		return 0, err
	}

	// TODO: 对不通情况的分数进行归一化处理
	// req == 1, score 表示 该卡到另外所有可用卡的带宽只和。越小越好
	// req == 2, score 表示 节点中 带宽最高的那两张卡之间的带宽。越大越好
	// req == 3, score 表示 节点中 3个卡两两相连接的3条相互连接的带宽和。越大越好
	// req == 4, score 表示 节点中 4个卡两两相连接的6条相互连接的带宽和。越大越好
	// ...
	if req == 1 {
		_, score, err := p.PreAllocate(n, req)
		if err != nil {
			return 0, err
		}

		return 1000 - score, nil
	}

	_, score, err := p.PreAllocate(n, req)
	return score, err
}

func (p *topologyPolicy) Allocate(n *NodeInfo, req int) ([]int, error) {
	ids, _, err := p.PreAllocate(n, req)
	return ids, err
}

// PreAllocate 计算分配方案，及该方案的打分
func (p *topologyPolicy) PreAllocate(n *NodeInfo, req int) (ids []int, score int, err error) {
	availableGPUs := n.getAvailableGPUs()
	if req <= 0 || req > availableGPUs {
		err = fmt.Errorf("rqu gpu count %v is invalid", req)
		return nil, 0, err
	}

	if req == 1 {
		minScoreId, score, err := getMinScoreGpu(n)
		if err != nil {
			return nil, 0, err
		}
		ids = append(ids, minScoreId)
		return ids, score, err
	}

	log.Printf("request gpus counts %v is more than 1, ", req)
	ids, score, err = getMaxScoreLink(n)
	if req == 2 {
		return ids, score, err
	}

	// 接下来是图中寻找最小生成树问题，参见 prim 算法

	// 标记 最大带宽已经使用
	n.devs[int(ids[0])].isUsed = true
	n.devs[int(ids[1])].isUsed = true

	// 计算接下来的卡到 集合ids 的带宽和，将最大到带宽和卡加入到 ids
	// 循环 req - 2此
	for c := 2; c < req; c++ {
		// 寻找到集合 d 最大的卡
		u := -1   // 记录当前遍历的卡
		max := -1 // 记录当前遍历的卡到集合的带宽

		for i := 0; i < len(n.devs); i++ {
			if n.devs[i].isUsed == false && calculateFromGPUAndSelectedSets(n, i, ids) > max {
				u = i
				max = calculateFromGPUAndSelectedSets(n, i, ids)
			}
		}

		if u == -1 { // 说明以上到遍历没有找到卡
			err = fmt.Errorf("rqu gpu count %v is invalidl", req)
			return ids, 0, err
		}

		// 将最高 离ids带宽和 的卡加入ids
		n.devs[u].isUsed = true
		ids = append(ids, u)
		score += max
	}

	return ids, score, nil
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

// 计算分数最高的边，返回两个节点, 可能多个分数最大值 idss,
func getMaxScoreLink(n *NodeInfo) (idss [][]int, score int, err error) {
	maxScore := -1
	for _, dev1 := range n.devs {
		if dev1.isUsed == false {
			for _, dev2 := range n.devs {
				if dev1 != dev2 && dev2.isUsed == false {
					score := calculateGPUPairScore(n, dev1.idx, dev2.idx)
					if score > maxScore {
						maxScore = score
						// 清空 idss
						idss = [][]int{}
						idss[0] = []int{dev1.idx, dev2.idx}
					} else if score == maxScore {
						idss = append(idss, []int{dev1.idx, dev2.idx})
					}
				}
			}
		}
	}

	if maxScore == -1 {
		err = fmt.Errorf("the node %s can't choose unused gpu", n.node.Name)
		return [][]int{}, 0, nil
	}

	return idss, maxScore, nil
}

// 计算一个候选 gpu 到已经确定的gpu集合的带宽和。
func calculateFromGPUAndSelectedSets(n *NodeInfo, candidate int, selectedSets []int) int {
	if len(selectedSets) == 0 {
		return 0
	}

	if candidate < 0 || candidate > len(n.devs) {
		return 0
	}

	score := 0
	for _, selectedSet := range selectedSets {
		if selectedSet == candidate {
			return 0
		}
		score += calculateGPUPairScore(n, candidate, selectedSet)
	}

	return score
}