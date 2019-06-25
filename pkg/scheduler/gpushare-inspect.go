package scheduler

import (
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/cache"
)

func NewGPUShareInspect(c *cache.SchedulerCache) *Inspect {
	return &Inspect{
		Name:  "gputopologyinspect",
		cache: c,
	}
}

type Result struct {
	Nodes []*Node `json:"nodes"`
	Error string  `json:"error,omitempty"`
}

type Node struct {
	Name        string                 `json:"name"`
	TotalGPU    uint                   `json:"totalGPU"`
	UsedGPU     uint                   `json:"usedGPU"`
	Devices     []*Device              `json:"devs"`
	GpuTopology map[uint]map[uint]uint `json:"gpuTopology"`
}

type Device struct {
	ID     int    `json:"id"`
	isUsed bool   `json:"usUsed""`
	Pods   []*Pod `json:"pods"`
}

type Pod struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	UsedGPU   int    `json:"usedGPU"`
}

type Inspect struct {
	Name  string
	cache *cache.SchedulerCache
}
