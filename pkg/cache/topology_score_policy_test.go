package cache

import (
	"testing"
)

func TestGetMaxScoreLink(t *testing.T) {
	devMap := map[int]*DeviceInfo{}
	gpuTopology := [][]TopologyType{
		{1, 1},
		{1, 2},
	}
	
	n := &NodeInfo{
		name:        "testnode",
		devs:        devMap,
		gpuCount:    8,
		gpuTopology: gpuTopology,
	}
	
	
	t.Log("hello world", n)
}


