package cache

import "fmt"

type Scheduler struct {
	node  *NodeInfo
	policy Policy
}

// NewAllocator creates a new Allocator using the given allocation policy
func NewScheduler(node *NodeInfo, policy Policy) (*Scheduler, error) {
	if policy == nil || node == nil {
		return nil, fmt.Errorf("new scheduler error policy in %v in node %v", policy, node)
	}
	return &Scheduler{
		node: node,
		policy: policy,
	}, nil
}

func (s *Scheduler) Score(request int) (int, error) {
	return s.policy.Score(s.node, request)
}

func (s *Scheduler) Allocate(request int) ([]int, error) {
	return s.policy.Allocate(s.node, request)
}