package types

import (
	"log"
	"sync"

	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/utils"
	"k8s.io/api/core/v1"
)

const (
	OptimisticLockErrorMsg = "the object has been modified; please apply your changes to the latest version and try again"
)

// NodeInfo is node level aggregateyd information.
type NodeInfo struct {
	name        string
	node        *v1.Node
	devs        map[int]*DeviceInfo
	gpuUsed     int
	gpuCount    int
	gpuTopology Topology
	rwmu        *sync.RWMutex
}

// Create Node Level
func NewNodeInfo(node *v1.Node) *NodeInfo {
	log.Printf("debug: NewNodeInfo() creates nodeInfo for %s", node.Name)

	devMap := map[int]*DeviceInfo{}
	for i := 0; i < utils.GetGPUCountInNode(node); i++ {
		devMap[i] = newDeviceInfo(i)
	}
	gpuTopology := NewTopology(node)

	nodeInfo := &NodeInfo{
		name:        node.Name,
		node:        node,
		devs:        devMap,
		gpuCount:    utils.GetGPUCountInNode(node),
		gpuTopology: gpuTopology,

		rwmu: new(sync.RWMutex),
	}

	log.Printf("debug: node %s has nodeinfo %v", node.Name, nodeInfo)
	log.Printf("debug: node %s has topology %v", node.Name, gpuTopology)
	return nodeInfo
}



func (n *NodeInfo) GetName() string {
	return n.name
}

func (n *NodeInfo) GetDevs() map[int]*DeviceInfo {
	return n.devs
}

func (n *NodeInfo) GetNode() *v1.Node {
	return n.node
}

func (n *NodeInfo) GetGPUCount() int {
	return n.gpuCount
}

func (n *NodeInfo) GetGPUTopology() [][]TopologyType {
	return n.gpuTopology
}

func (n *NodeInfo) GetGPUUsedCount() int {
	count := 0
	for _, dev := range n.devs {
		if dev.isUsed {
			count++
		}
	}
	return count
}

func (n *NodeInfo) RemovePod(pod *v1.Pod) {
	n.rwmu.Lock()
	defer n.rwmu.Unlock()

	ids := utils.GetGPUIDFromAnnotation(pod)
	originTopology := getOriginGpuTopology(n.node)
	log.Printf("warn: Pod remove ids %v", ids)
	for _, id := range ids {
		if id >= 0 {
			dev, found := n.devs[id]
			if !found {
				log.Printf("warn: Pod %s in ns %s failed to find the GPU ID %d in node %s", pod.Name, pod.Namespace, id, n.name)
			} else {
				dev.RemovePod(pod)
				n.gpuTopology.RecoveryId(id, originTopology)
			}
		} else {
			log.Printf("warn: Pod %s in ns %s is not set the GPU ID %d in node %s", pod.Name, pod.Namespace, id, n.name)
		}
	}
}

// Add the Pod which has the GPU id to the node
func (n *NodeInfo) AddOrUpdatePod(pod *v1.Pod) (added bool) {
	n.rwmu.Lock()
	defer n.rwmu.Unlock()

	ids := utils.GetGPUIDFromAnnotation(pod)
	log.Printf("debug: addOrUpdatePod() Pod %s in ns %s with the GPU IDs %v should be added to device map",
		pod.Name,
		pod.Namespace,
		ids)
	if len(ids) > 0 {
		for _, id := range ids {
			dev, found := n.devs[id]
			if !found {
				log.Printf("warn: Pod %s in ns %s failed to find the GPU ID %d in node %s", pod.Name, pod.Namespace, id, n.name)
			} else {
				dev.AddPod(pod)
				n.gpuTopology.ConsumeId(id)
				added = true
			}
		}
	} else {
		log.Printf("warn: Pod %s in ns %s is not set the GPU ID in node %s", pod.Name, pod.Namespace, n.name)
	}
	return added
}

// check if the pod can be allocated on the node
func (n *NodeInfo) Assume(pod *v1.Pod) (allocatable bool) {
	allocatable = false

	n.rwmu.RLock()
	defer n.rwmu.RUnlock()

	availableGPUs := n.GetAvailableGPUs()
	reqGPU := utils.GetGPUCountFromPodResource(pod)
	log.Printf("debug: AvailableGPUs: %v in node %s", availableGPUs, n.name)
	log.Printf("debug: requestGPUs: %v in node %s", reqGPU, n.name)

	if availableGPUs > 0 && availableGPUs-reqGPU >= 0 {
		allocatable = true
	}

	return
}

func (n *NodeInfo) GetAvailableGPUs() (availableGPUs int) {
	allGPUs := n.GetAllGPUs()
	usedGPUs := n.GetGPUUsedCount()
	availableGPUs = allGPUs - usedGPUs
	return availableGPUs
}

func (n *NodeInfo) GtUsedGPUs() (usedGPUs int) {
	usedGPUs = n.GetGPUUsedCount()
	log.Printf("debug: getUsedGPUs: %v in node %s, and devs %v", usedGPUs, n.name, n.devs)
	return usedGPUs
}

func (n *NodeInfo) GetAllGPUs() (allGPUs int) {
	allGPUs = n.GetGPUCount()
	log.Printf("debug: getAllGPUs: %v in node %s, and dev %v", allGPUs, n.name, n.devs)
	return allGPUs
}
