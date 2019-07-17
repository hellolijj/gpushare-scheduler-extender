package cache

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/utils"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
)

const (
	OptimisticLockErrorMsg = "the object has been modified; please apply your changes to the latest version and try again"
)

// NodeInfo is node level aggregated information.
type NodeInfo struct {
	name        string
	node        *v1.Node
	devs        map[int]*DeviceInfo
	gpuUsed     int
	gpuCount    int
	gpuTopology [][]TopologyType
	rwmu        *sync.RWMutex
}

type TopologyType nvml.P2PLinkType

// Create Node Level
func NewNodeInfo(node *v1.Node) *NodeInfo {
	log.Printf("debug: NewNodeInfo() creates nodeInfo for %s", node.Name)

	devMap := map[int]*DeviceInfo{}
	for i := 0; i < utils.GetGPUCountInNode(node); i++ {
		devMap[i] = newDeviceInfo(i) // 这里的i 表示什么？device id 吗？
	}

	gpuTopology := getGPUTopologyFromNode(node, devMap)

	nodeInfo := &NodeInfo{
		name:        node.Name,
		node:        node,
		devs:        devMap,
		gpuCount:    utils.GetGPUCountInNode(node),
		gpuTopology: gpuTopology,

		rwmu: new(sync.RWMutex),
	}

	log.Printf("debug: node %s has nodeinfo %v", node.Name, nodeInfo)
	return nodeInfo
}

// 从node annotaion 里获取gpu topology
func getGPUTopologyFromNode(node *v1.Node, devs map[int]*DeviceInfo) [][]TopologyType {
	// init gpuTopology
	topology := make([][]TopologyType, len(devs))

	if !utils.IsGPUTopologyNode(node) {
		return topology
	}

	for i := 0; i < len(devs); i++ {
		topology[i] = make([]TopologyType, len(devs))
	}

	for k, v := range node.Annotations {
		// GPU_SYS_0_1: Cross CPU socket
		if strings.HasPrefix(k, utils.GPU_PRIFX) {
			var gpu1, gpu2 int
			var topoAbbr, topoDesc string
			fmt.Sscanf(k, utils.GPU_PRIFX+"%s_%d_%d", &topoAbbr, &gpu1, &gpu2)
			fmt.Sscanf(v, "%s", &topoDesc)
			topology[gpu1][gpu2] = TopologyType(utils.GetGPULinkFromDescAndAbbr(topoDesc, topoAbbr))
			topology[gpu2][gpu1] = TopologyType(utils.GetGPULinkFromDescAndAbbr(topoDesc, topoAbbr))
		}
	}

	return topology
}

func (n *NodeInfo) GetName() string {
	return n.name
}

func (n *NodeInfo) GetDevs() []*DeviceInfo {
	devs := make([]*DeviceInfo, n.gpuCount)
	for i, dev := range n.devs {
		devs[i] = dev
	}
	return devs
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

func (n *NodeInfo) removePod(pod *v1.Pod) {
	n.rwmu.Lock()
	defer n.rwmu.Unlock()

	ids := utils.GetGPUIDFromAnnotation(pod)
	log.Printf("warn: Pod remove ids %v", ids)
	for _, id := range ids {
		if id >= 0 {
			dev, found := n.devs[id]
			if !found {
				log.Printf("warn: Pod %s in ns %s failed to find the GPU ID %d in node %s", pod.Name, pod.Namespace, id, n.name)
			} else {
				dev.removePod(pod)
			}
		} else {
			log.Printf("warn: Pod %s in ns %s is not set the GPU ID %d in node %s", pod.Name, pod.Namespace, id, n.name)
		}
	}
}

// Add the Pod which has the GPU id to the node
func (n *NodeInfo) addOrUpdatePod(pod *v1.Pod) (added bool) {
	n.rwmu.Lock()
	defer n.rwmu.Unlock()

	ids := utils.GetGPUIDFromAnnotation(pod)
	log.Printf("debug: addOrUpdatePod() Pod %s in ns %s with the GPU IDs %v should be added to device map",
		pod.Name,
		pod.Namespace,
		ids)
	if len(ids) >= 0 {
		for _, id := range ids {
			dev, found := n.devs[id]
			if !found {
				log.Printf("warn: Pod %s in ns %s failed to find the GPU ID %d in node %s", pod.Name, pod.Namespace, id, n.name)
			} else {
				dev.addPod(pod)
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

	availableGPUs := n.getAvailableGPUs()
	reqGPU := utils.GetGPUCountFromPodResource(pod)
	log.Printf("debug: AvailableGPUs: %v in node %s", availableGPUs, n.name)
	log.Printf("debug: requestGPUs: %v in node %s", reqGPU, n.name)

	if availableGPUs > 0 && availableGPUs-reqGPU >= 0 {
		allocatable = true
	}

	return
}

func (n *NodeInfo) Allocate(clientset *kubernetes.Clientset, pod *v1.Pod) (err error) {
	var newPod *v1.Pod
	n.rwmu.Lock()
	defer n.rwmu.Unlock()
	log.Printf("debug: Allocate() ----Begin to allocate GPU for gpu topology for pod %s in ns %s----", pod.Name, pod.Namespace)
	// 1. Update the pod spec
	devIds, found := n.allocateGPUID(pod)
	if found {
		log.Printf("debug: Allocate() 1. Allocate GPU ID %v to pod %s in ns %s.----", devIds, pod.Name, pod.Namespace)
		// newPod := utils.GetUpdatedPodEnvSpec(pod, devId, nodeInfo.GetTotalGPUMemory()/nodeInfo.GetGPUCount())
		newPod = utils.GetUpdatedPodAnnotationSpec(pod, devIds)
		_, err = clientset.CoreV1().Pods(newPod.Namespace).Update(newPod)
		if err != nil {
			// the object has been modified; please apply your changes to the latest version and try again
			if err.Error() == OptimisticLockErrorMsg {
				// retry
				pod, err = clientset.CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
				if err != nil {
					return err
				}
				// newPod = utils.GetUpdatedPodEnvSpec(pod, devId, nodeInfo.GetTotalGPUMemory()/nodeInfo.GetGPUCount())
				newPod = utils.GetUpdatedPodAnnotationSpec(pod, devIds)
				_, err = clientset.CoreV1().Pods(newPod.Namespace).Update(newPod)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
	} else {
		err = fmt.Errorf("the node %s can't place the pod %s in ns %s", pod.Spec.NodeName, pod.Name, pod.Namespace)
	}

	// 2. Bind the pod to the node
	if err == nil {
		binding := &v1.Binding{
			ObjectMeta: metav1.ObjectMeta{Name: pod.Name, UID: pod.UID},
			Target:     v1.ObjectReference{Kind: "Node", Name: n.name},
		}
		log.Printf("debug: Allocate() 2. Try to bind pod %s in %s namespace to node %s with %v",
			pod.Name,
			pod.Namespace,
			pod.Spec.NodeName,
			binding)
		err = clientset.CoreV1().Pods(pod.Namespace).Bind(binding)
		if err != nil {
			log.Printf("warn: Failed to bind the pod %s in ns %s due to %v", pod.Name, pod.Namespace, err)
			return err
		}
	}

	// 3. update the device info if the pod is update successfully
	if err == nil {
		log.Printf("debug: Allocate() 3. Try to add pod %s in ns %s to dev %v",
			pod.Name,
			pod.Namespace,
			devIds)
		for _, devId := range devIds {
			dev, found := n.devs[int(devId)]
			if !found {
				log.Printf("warn: Pod %s in ns %s failed to find the GPU ID %d in node %s", pod.Name, pod.Namespace, devId, n.name)
			} else {
				dev.addPod(newPod)
			}
		}
	}
	log.Printf("debug: Allocate() ----End to allocate GPU for gpu mem for pod %s in ns %s----", pod.Name, pod.Namespace)
	return err
}

// allocate the GPU ID to the pod
func (n *NodeInfo) allocateGPUID(pod *v1.Pod) (candidateDevID []uint, found bool) {
	reqGPU := 0
	found = false
	availableGPUs := n.getAvailableGPUs()
	reqGPU = utils.GetGPUCountFromPodResource(pod)
	log.Printf("debug: reqGPU for pod %s in ns %s: %d", pod.Name, pod.Namespace, reqGPU)
	log.Printf("debug: AvailableGPUs: %v in node %s", availableGPUs, n.name)
	
	topologyScheduler, err := NewScheduler(n, NewTopologyPolicy())
	if err != nil {
		log.Printf("warn: Failed to get scheduler object %v", topologyScheduler)
		return
	}

	if reqGPU > 0 {
		if availableGPUs > 0 && availableGPUs-reqGPU >= 0 {
			
			ids, err := topologyScheduler.policy.Allocate(n, reqGPU)
			if err != nil {
				log.Printf("allocate gpu to node failed, resaon: %v", err)
				return
			}
			for _, id := range ids {
				candidateDevID = append(candidateDevID, uint(id))
			}
			found = true
		}
		if found {
			log.Printf("debug: Find candidate dev id %d for pod %s in ns %s successfully.",
				candidateDevID,
				pod.Name,
				pod.Namespace)
		} else {
			log.Printf("warn: Failed to find available GPUs %d for the pod %s in the namespace %s",
				reqGPU,
				pod.Name,
				pod.Namespace)
		}
	}

	return candidateDevID, found
}

func (n *NodeInfo) getAvailableGPUs() (availableGPUs int) {
	allGPUs := n.getAllGPUs()
	usedGPUs := n.getUsedGPUs()
	availableGPUs = allGPUs - usedGPUs
	return availableGPUs
}

func (n *NodeInfo) getUsedGPUs() (usedGPUs int) {
	usedGPUs = n.GetGPUUsedCount()
	log.Printf("debug: getUsedGPUs: %v in node %s, and devs %v", usedGPUs, n.name, n.devs)
	return usedGPUs
}

func (n *NodeInfo) getAllGPUs() (allGPUs int) {
	allGPUs = n.GetGPUCount()
	log.Printf("debug: getAllGPUs: %v in node %s, and dev %v", allGPUs, n.name, n.devs)
	return allGPUs
}