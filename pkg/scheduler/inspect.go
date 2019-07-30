package scheduler

import (
	"log"

	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/cache"
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/policy"
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/types"
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/utils"
)

type Inspect struct {
	Name   string
	cache  *cache.SchedulerCache
	policy *policy.Policy
}

func (in Inspect) Handler(name string, detail bool) *types.InspectResult {
	nodes := []*types.InspectNode{}
	errMsg := ""

	if len(name) == 0 {
		nodeInfos := in.cache.ListNodeInfo()
		for _, info := range nodeInfos {
			nodes = append(nodes, in.buildNode(info, detail))
		}

	} else {
		node, err := in.cache.GetNodeInfo(name)
		if err != nil {
			errMsg = err.Error()
		}
		if len(node.GetName()) == 0 {
			errMsg = "invalid node name"
		}
		nodes = append(nodes, in.buildNode(node, detail))
	}

	log.Printf("debug: policy name %v", in.policy.GetName())

	return &types.InspectResult{
		Nodes:  nodes,
		Policy: in.policy.GetName(),
		Error:  errMsg,
	}
}

func (in Inspect) buildNode(info *types.NodeInfo, detail bool) *types.InspectNode {
	if !detail {
		return &types.InspectNode{
			Name:     info.GetName(),
			TotalGPU: info.GetGPUCount(),
			UsedGPU:  info.GetGPUUsedCount(),
		}
	}

	topology := info.GetGPUTopology()
	log.Printf("debug: inspect origin topology: %v", topology)

	// to get static config
	// staticSet := policy.NodeTypeConfig()[utils.GetNodeTypeFromAnnotation(info.GetNode())]

	return &types.InspectNode{
		Name:     info.GetName(),
		TotalGPU: info.GetGPUCount(),
		UsedGPU:  info.GetGPUUsedCount(),
		Topology: info.GetGPUTopology(),
		NodeType: utils.GetNodeTypeFromAnnotation(info.GetNode()),
	}
}
