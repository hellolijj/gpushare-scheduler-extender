package cache

import (
	"log"
	"sync"
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/utils"
	"k8s.io/api/core/v1"
	
	
	gputype "github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/types"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	corelisters "k8s.io/client-go/listers/core/v1"
)

type SchedulerCache struct {

	// a map from pod key to podState.
	nodes map[string]*gputype.NodeInfo

	// nodeLister can list/get nodes from the shared informer's store.
	nodeLister corelisters.NodeLister

	//
	podLister corelisters.PodLister

	// record the knownPod, it will be added when annotation ALIYUN_GPU_ID is added, and will be removed when complete and deleted
	knownPods map[types.UID]*v1.Pod
	nLock     *sync.RWMutex
}

func NewSchedulerCache(nLister corelisters.NodeLister, pLister corelisters.PodLister) *SchedulerCache {
	return &SchedulerCache{
		nodes:      make(map[string]*gputype.NodeInfo),
		nodeLister: nLister,
		podLister:  pLister,
		knownPods:  make(map[types.UID]*v1.Pod),
		nLock:      new(sync.RWMutex),
	}
}

func (cache *SchedulerCache) ListNodeInfo() []*gputype.NodeInfo {
	nodes := []*gputype.NodeInfo{}
	log.Println("debug:list nodes info %v", cache.nodes)
	for _, n := range cache.nodes {
		nodes = append(nodes, n)
	}
	return nodes
}

// build cache when initializing
func (cache *SchedulerCache) BuildCache() error {
	log.Println("debug: begin to build scheduler pod cache")
	pods, err := cache.podLister.List(labels.Everything())

	if err != nil {
		return err
	} else {
		for _, pod := range pods {
			if utils.GetGPUCountFromPodAnnotation(pod) <= 0 {
				continue
			}

			if len(pod.Spec.NodeName) == 0 {
				continue
			}

			err = cache.AddOrUpdatePod(pod)
			
			if err != nil {
				return err
			}
		}

	}
	
	log.Println("debug: begin to build scheduler node cache")
	nodes, err := cache.nodeLister.List(labels.Everything())
	
	if err != nil {
		return err
	} else {
		for _, node := range nodes {
			if len(node.Name) == 0 {
				continue
			}
			if !utils.IsGPUTopologyNode(node) {
				continue
			}
			_, ok := cache.nodes[node.Name]
			if ok {
				continue
			}
			
			cache.nodes[node.Name] = gputype.NewNodeInfo(node)
		}
	}

	return nil
}

func (cache *SchedulerCache) GetPod(name, namespace string) (*v1.Pod, error) {
	return cache.podLister.Pods(namespace).Get(name)
}

// Get known pod from the pod UID
func (cache *SchedulerCache) KnownPod(podUID types.UID) bool {
	cache.nLock.RLock()
	defer cache.nLock.RUnlock()

	_, found := cache.knownPods[podUID]
	return found
}

func (cache *SchedulerCache) AddOrUpdatePod(pod *v1.Pod) error {
	log.Printf("debug: Add or update pod info: %v", pod.Name)
	log.Printf("debug: Node %v", cache.nodes)
	if len(pod.Spec.NodeName) == 0 {
		log.Printf("debug: pod %s in ns %s is not assigned to any node, skip", pod.Name, pod.Namespace)
		return nil
	}
	
	n, err := cache.GetNodeInfo(pod.Spec.NodeName)
	if err != nil {
		return err
	}
	podCopy := pod.DeepCopy()
	if n.AddOrUpdatePod(podCopy) {
		// put it into known pod
		cache.rememberPod(pod.UID, podCopy)
	} else {
		log.Printf("debug: pod %s in ns %s's gpu id is %d, it's illegal, skip",
			pod.Name,
			pod.Namespace,
			utils.GetGPUIDFromAnnotation(pod))
	}

	return nil
}

// The lock is in cacheNode
func (cache *SchedulerCache) RemovePod(pod *v1.Pod) {
	log.Printf("debug: Node %v", cache.nodes)
	n, err := cache.GetNodeInfo(pod.Spec.NodeName)
	if err == nil {
		log.Printf("debug: Remove pod info: %v", pod.Name)
		n.RemovePod(pod)
	} else {
		log.Printf("debug: Failed to get node %s due to %v", pod.Spec.NodeName, err)
	}

	cache.forgetPod(pod.UID)
}

// Get or build nodeInfo if it doesn't exist
func (cache *SchedulerCache) GetNodeInfo(name string) (*gputype.NodeInfo, error) {
	node, err := cache.nodeLister.Get(name)
	if err != nil {
		return nil, err
	}
	
	log.Printf("debug: cache nodes %v", cache.nodes)

	cache.nLock.Lock()
	defer cache.nLock.Unlock()
	n, ok := cache.nodes[name]

	if !ok {
		n = gputype.NewNodeInfo(node)
		cache.nodes[name] = n
	} else {
		// if the existing node turn from non gpu to gpu
		if utils.GetGPUCountInNode(n.GetNode()) <= 0 && utils.GetGPUCountInNode(node) > 0 {
			log.Printf("debug: GetNodeInfo() need update node %s from %v to %v",
				name,
				n.GetNode(),
				node)
			n = gputype.NewNodeInfo(node)
			cache.nodes[name] = n
		}

		log.Printf("debug: GetNodeInfo() uses the existing nodeInfo for %s", name)
	}
	return n, nil
}

func (cache *SchedulerCache) forgetPod(uid types.UID) {
	cache.nLock.Lock()
	defer cache.nLock.Unlock()
	delete(cache.knownPods, uid)
}

func (cache *SchedulerCache) rememberPod(uid types.UID, pod *v1.Pod) {
	cache.nLock.Lock()
	defer cache.nLock.Unlock()
	cache.knownPods[pod.UID] = pod
}
