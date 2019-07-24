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

type NodeType int
// Valid GPUTypes
const (
	UnknownNodeType NodeType = iota
	ShenLongNode
)

var (
	NodeTypeMap  = map[string]NodeType{
		NodeTypeSHENGLONG: ShenLongNode,
	}
)
// 分装节点烈性
func GetNodeTypeFromAnnotation(node *v1.Node) NodeType {
	nodeTypeKey, ok := node.Annotations[EnvNodeType]
	if !ok {
		return UnknownNodeType
	}
	
	nodeType, ok := NodeTypeMap[nodeTypeKey]
	if !ok {
		return UnknownNodeType
	}
	
	return nodeType
}