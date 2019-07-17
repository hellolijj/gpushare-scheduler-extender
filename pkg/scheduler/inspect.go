package scheduler

func (in Inspect) Handler(name string) *Result {
	nodes := []*Node{}
	errMsg := ""
	
	return &Result{
		Nodes: nodes,
		Error: errMsg,
	}
}
