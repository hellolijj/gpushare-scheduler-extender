package cache

// gpu edge single gpu is not include
type Edge struct {
	gpu1     uint
	gpu2     uint
	distance uint
}

type Edges []*Edge

func (e Edges) Len() int {
	return len(e)
}

func (e Edges) Less(i, j int) bool {
	return e[i].distance < e[j].distance
}

func (e Edges) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
