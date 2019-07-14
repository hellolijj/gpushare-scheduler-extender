package cache


type simplePolicy struct{}

// NewSimplePolicy creates a new SimplePolicy.
func NewSimplePolicy() Policy {
	return &simplePolicy{}
}

type Policy interface {
	Score(node *NodeInfo, quest int) int
}

// Allocate GPUs following a simple policy.
func (p *simplePolicy) Score(node *NodeInfo, ques int) int {
	if ques <= 0 {
		return 0
	}
	return 3
}
