package policy

import (
	"log"

	
	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/types"
)


// 计算两块 gpu 之间的分数
func calculateGPUPairScore(n *types.NodeInfo, gpu0 int, gpu1 int) int {
	if gpu0 < 0 || gpu0 > len(n.GetDevs())-1 || gpu1 < 0 || gpu1 > len(n.GetDevs())-1 {
		log.Printf("invaild gpu pair format %d-%d", gpu0, gpu1)
		return 0
	}

	if gpu0 == gpu1 {
		log.Printf("can't calculate the same gpu score: %v", gpu1)
		return 0
	}
	
	devs := n.GetDevs()

	if devs[gpu0].IsUsed() || devs[gpu1].IsUsed() {
		log.Printf("gpu can't be used: gpu%d %v, gpu%d, %v", gpu0, devs[gpu0].IsUsed(), gpu1, devs[gpu1].IsUsed())
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
