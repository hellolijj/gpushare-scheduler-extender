package cache

import "fmt"

type simplePolicy struct{}

// NewSimplePolicy creates a new SimplePolicy.
func NewSimplePolicy() Policy {
	return &simplePolicy{}
}

type Policy interface {
	Score(node *NodeInfo, quest int) (int, error)
	Allocate(n *NodeInfo, req int) ([]int, error)
}

// 所有节点的打分都是一样的
func (p *simplePolicy) Score(node *NodeInfo, ques int) (int, error) {
	if ques <= 0 {
		return 0, nil
	}

	return 3, nil
}

func (p *simplePolicy) Allocate(n *NodeInfo, req int) ([]int, error) {
	availableGPUs := n.getAvailableGPUs()
	if req <= 0 || req > availableGPUs {
		err := fmt.Errorf("rqu gpu count %v is invalid", req)
		return []int{}, err
	}
	
	ids := []int{}
	for i := 0; i < req; i++ {
		if n.devs[i].isUsed == false {
			ids = append(ids, i)
		}
	}
	return ids, nil
}
