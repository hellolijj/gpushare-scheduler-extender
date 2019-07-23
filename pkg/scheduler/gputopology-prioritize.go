package scheduler

import (
	"log"

	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/cache"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/utils"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
)

func NewGPUTopologyPrioritize(clientset *kubernetes.Clientset, c *cache.SchedulerCache) *Prioritize {
	return &Prioritize{
		Name: "gputopologysort",
		Func: func(pod v1.Pod, nodes []v1.Node, c *cache.SchedulerCache) (*schedulerapi.HostPriorityList, error) {
			log.Printf("debug:  gpu topology prioritize in nodes: %v", nodes)
			var priorityList schedulerapi.HostPriorityList
			priorityList = make([]schedulerapi.HostPriority, len(nodes))
			reqGPU := utils.GetGPUCountFromPodResource(&pod)

			for i, node := range nodes {

				nodeInfo, err := c.GetNodeInfo(node.Name)
				if err != nil {
					log.Printf("warn: Failed to handle node %s in ns %s due to error %v", node.Name, node.Namespace, err)
					return &priorityList, err
				}
				
				// topologyScheduler, err := cache.NewScheduler(nodeInfo, cache.NewTopologyPolicy())
				bestScheduler, err := cache.NewScheduler(nodeInfo, cache.NewBestPolicy())
				if err != nil {
					log.Printf("warn: Failed to get scheduler object %v", bestScheduler)
					return &priorityList, err
				}

				// here to sort in node
				score, err := bestScheduler.Score(reqGPU)
				if err != nil {
					log.Printf("warn: Failed to score in node %s in ns %s due to error %v", node.Name, node.Namespace, err)
					return &priorityList, err
				}
				
				priorityList[i] = schedulerapi.HostPriority{
					Host:  node.Name,
					Score: score,
				}
			}
			return &priorityList, nil
		},
		cache: c,
	}
}
