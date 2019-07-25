package policy

import (
	"fmt"

	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/utils"
)

type simpleRunner struct{}

// NewSimplePolicy creates a new SimplePolicy.
func NewSimpleRunner() Run {
	return &simpleRunner{}
}

// 所有节点的打分都是一样的
func (s *simpleRunner) Score(n *utils.NodeInfo, ques int) (int, error) {
	if ques <= 0 {
		return 0, nil
	}
	return 10, nil
}

func (s *simpleRunner) Allocate(n *utils.NodeInfo, req int) ([]int, error) {
	availableGPUs := n.GetAvailableGPUs()
	if req <= 0 || req > availableGPUs {
		err := fmt.Errorf("rqu gpu count %v is invalid", req)
		return []int{}, err
	}

	ids := []int{}
	devs := n.GetDevs()
	for i := 0; i < req; i++ {
		if devs[i].IsUsed() == false {
			ids = append(ids, i)
		}
	}
	return ids, nil

}
