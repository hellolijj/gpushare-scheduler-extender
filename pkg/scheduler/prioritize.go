package scheduler

import (
	"k8s.io/api/core/v1"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
	
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/cache"
)


type Prioritize struct {
	Name string
	Func func(pod v1.Pod, nodes []v1.Node, c *cache.SchedulerCache) (*schedulerapi.HostPriorityList, error)
	cache *cache.SchedulerCache
	
}

func (p Prioritize) Handler(args schedulerapi.ExtenderArgs) (*schedulerapi.HostPriorityList, error) {
	return p.Func(*args.Pod, args.Nodes.Items, p.cache)
}