package policy

import (
	"sync"

	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/utils"
)

type Policy struct {
	name   string
	config string
	Run    Run
	rwmu   *sync.RWMutex
}

// NewAllocator creates a new Allocator using the given allocation policy
func NewPolicy(name, config string) (*Policy, error) {
	return &Policy{
		name:   name,
		config: config,
		Run:    newPolicyRuner(name, config),
		rwmu:   new(sync.RWMutex),
	}, nil
}

func newPolicyRuner(name, config string) Run {
	switch name {
	case "simple":
		return NewSimpleRun()
	case "base_effort":
		return NewBestRun()
	default:
	}
	return nil
}

/*
func (p *Policy) Score(request int) (int, error) {
	return p.run.Score()
}

func (s *Scheduler) Allocate(request int) ([]int, error) {
	return s.policy.Allocate(s.node, request)
}
*/

type Run interface {
	Score(n *utils.NodeInfo, req int) (int, error)
	Allocate(n *utils.NodeInfo, req int) ([]int, error)
}

func (p *Policy)Score(n *utils.NodeInfo, req int) (int, error) {
	p.rwmu.Lock()
	defer p.rwmu.Unlock()
	return p.Run.Score(n, req)
}