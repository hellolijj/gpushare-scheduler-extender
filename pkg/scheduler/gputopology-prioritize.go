package scheduler

import (
	"log"

	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/cache"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
)

func NewGPUTopologyPrioritize(clientset *kubernetes.Clientset, c *cache.SchedulerCache) *Prioritize {
	return &Prioritize{
		Name: "gputopologysort",
		Func: func(pod v1.Pod, nodes []v1.Node, c *cache.SchedulerCache) (*schedulerapi.HostPriorityList, error) {
			log.Printf("debug:  gpu topology prioritize in nodes: %v", nodes)
			var priorityList schedulerapi.HostPriorityList
			priorityList = make([]schedulerapi.HostPriority, len(nodes))
			for i, node := range nodes {
				
				// here to sort in node
				priorityList[i] = schedulerapi.HostPriority{
					Host:  node.Name,
					Score: 10,
				}
			}
			return &priorityList, nil
		},
		cache: c,
	}
}
