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

// 获取 annotation 上的node
func GetNodeTypeFromAnnotation(node *v1.Node) string {
	nodeType, ok := node.Annotations[EnvNodeType]
	if !ok {
		return ""
	}

	return nodeType
}
