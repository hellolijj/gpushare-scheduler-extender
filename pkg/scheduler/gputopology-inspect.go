package scheduler

import (
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/cache"
)

func NewGPUShareInspect(c *cache.SchedulerCache) *Inspect {
	return &Inspect{
		Name:  "gpushareinspect",
		cache: c,
	}
}

type Result struct {
	Nodes []*Node `json:"nodes"`
	Error string  `json:"error,omitempty"`
}

type Node struct {
	Name       string    `json:"name"`
	TotalGPU   int       `json:"totalGPU"`
	UsedGPU    int       `json:"usedGPU"`
	Topology [][]cache.TopologyType    `json:"topology"`
}


type Inspect struct {
	Name  string
	cache *cache.SchedulerCache
}
