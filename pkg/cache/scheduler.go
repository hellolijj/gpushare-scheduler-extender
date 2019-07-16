package cache

type Scheduler struct {
	node  *NodeInfo
	policy Policy
}

// NewAllocator creates a new Allocator using the given allocation policy
func NewScheduler(node *NodeInfo, policy Policy) (*Scheduler, error) {
	return &Scheduler{
		node: node,
		policy: policy,
	}, nil
}


func (a *Scheduler) Score(request int) (int, error) {
	return a.policy.Score(a.node, request)
}