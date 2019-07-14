package cache

type Scorer struct {
	node  *NodeInfo
	policy Policy
}

// NewAllocator creates a new Allocator using the given allocation policy
func NewScorer(node *NodeInfo, policy Policy) (*Scorer, error) {
	return &Scorer{
		node: node,
		policy: policy,
	}, nil
}


func (a *Scorer) Score(request int) int {
	return a.policy.Score(a.node, request)
}