package utils

import (
	"fmt"
	"strings"

	"k8s.io/api/core/v1"
)

// Is the Node for GPU Topology
func IsGPUTopologyNode(node *v1.Node) bool {
	return GetGPUCountInNode(node) > 0
}

// Get the total GPU memory of the Node
func GetTotalGPUMemory(node *v1.Node) int {
	val, ok := node.Status.Capacity[ResourceName]

	if !ok {
		return 0
	}

	return int(val.Value())
}

// Get the GPU count of the node
func GetGPUCountInNode(node *v1.Node) int {
	val, ok := node.Status.Capacity[ResourceName]

	if !ok {
		return int(0)
	}

	return int(val.Value())
}

func GetGPUTopologyInNode(node *v1.Node) map[uint]map[uint]uint {
	topology := make(map[uint]map[uint]uint)
	if !IsGPUTopologyNode(node) {
		return topology
	}

	for k, v := range node.Annotations {
		if strings.HasPrefix(k, "GSOC_DEV_") {
			var gpu1, gpu2, topo uint
			fmt.Sscanf(k, "GSOC_DEV_%d_%d", &gpu1, &gpu2)
			fmt.Sscanf(v, "%d", &topo)
			topology[gpu1] = map[uint]uint{gpu2: topo}
		}
	}
	
	return topology
}
