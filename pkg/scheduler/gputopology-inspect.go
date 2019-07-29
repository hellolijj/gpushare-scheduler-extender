package scheduler

import (
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/cache"
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/policy"
)

func NewGPUTopologyInspect(c *cache.SchedulerCache, policy *policy.Policy) *Inspect {
	return &Inspect{
		Name:  "gputopologyinspect",
		cache: c,
		policy: policy,
	}
}
