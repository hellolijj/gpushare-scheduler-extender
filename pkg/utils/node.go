package utils

import (
	"k8s.io/api/core/v1"
)

// Is the Node for GPU Topology
func IsGPUTopologyNode(node *v1.Node) bool {
	return GetGPUCountInNode(node) > 0
}

// Get the GPU count of the node
func GetGPUCountInNode(node *v1.Node) int {
	val, ok := node.Status.Capacity[ResourceName]

	if !ok {
		return int(0)
	}

	return int(val.Value())
}