package scheduler

import (
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/cache"
)

func (in Inspect) Handler(name string) *Result {
	nodes := []*Node{}
	errMsg := ""
	
	if len(name) == 0 {
		nodeInfos := in.cache.GetNodeinfos()
		for _, info := range nodeInfos {
			nodes = append(nodes, buildNode(info))
		}
		
	} else {
		node, err := in.cache.GetNodeInfo(name)
		if err != nil {
			errMsg = err.Error()
		}
		if len(node.GetName()) == 0 {
			errMsg = "invalid node name"
		}
		nodes = append(nodes, buildNode(node))
	}
	
	return &Result{
		Nodes: nodes,
		Error: errMsg,
	}
}

func buildNode(info *cache.NodeInfo) *Node {
	
	topology := info.GetGPUTopology()
	
	for _, dev := range info.GetDevs() {
		if dev.IsUsed() {
			// 所有与devid相关的拓扑都为-1
			for gpu0, topo := range topology {
				for gpu1, _ := range topo {
					if gpu0 == dev.GetDevId() || gpu1 == dev.GetDevId() {
						topology[gpu0][gpu1] = 0      // here -1 may be more right
					}
				}
			}
		}
	}
	
	return &Node{
		Name:     info.GetName(),
		TotalGPU: info.GetGPUCount(),
		UsedGPU	: info.GetGPUUsedCount(),
		Topology: topology,
	}
}