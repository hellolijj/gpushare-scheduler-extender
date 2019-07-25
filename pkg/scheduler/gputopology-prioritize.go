package scheduler

import (
	"log"

	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/cache"
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/policy"
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/utils"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
)

func NewGPUTopologyPrioritize(clientset *kubernetes.Clientset, c *cache.SchedulerCache, policy *policy.Policy) *Prioritize {
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

				// here to sort in node
				score, err := policy.Score(nodeInfo, reqGPU)
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
