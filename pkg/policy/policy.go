package policy

import (
	"sync"

	"fmt"
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/types"
)

type Policy struct {
	name   string
	config string
	Run    Run
	rwmu   *sync.RWMutex
}

// NewAllocator creates a new Allocator using the given allocation policy
func NewPolicy(name, config string) (*Policy, error) {
	runnerPolicy := newPolicyRunner(name, config)
	if runnerPolicy == nil || len(name) == 0 {
		return nil, fmt.Errorf("error in policy %v", name)
	}
	
	return &Policy{
		name:   name,
		config: config,
		Run:    runnerPolicy,
		rwmu:   new(sync.RWMutex),
	}, nil
}

func newPolicyRunner(name, config string) Run {
	switch name {
	case "simple":
		return NewSimpleRunner()
	case "best_effort":
		return NewBestRunner()
	case "static":
		return NewStaticRunner()
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
	Score(n *types.NodeInfo, req int) (int, error)
	Allocate(n *types.NodeInfo, req int) ([]int, error)
}

func (p *Policy)Score(n *types.NodeInfo, req int) (int, error) {
	p.rwmu.Lock()
	defer p.rwmu.Unlock()
	return p.Run.Score(n, req)
}