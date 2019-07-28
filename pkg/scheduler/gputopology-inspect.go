package scheduler

import (
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/cache"
)

func NewGPUTopologyInspect(c *cache.SchedulerCache) *Inspect {
	return &Inspect{
		Name:  "gputopologyinspect",
		cache: c,
	}
}


type Inspect struct {
	Name  string
	cache *cache.SchedulerCache
}
