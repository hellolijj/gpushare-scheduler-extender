package types

// todo 将所有 type 文件移到这里
// 还有将项目名称改过来，改为 helo


type InspectResult struct {
	Nodes []*InspectNode `json:"nodes"`
	Error string  `json:"error,omitempty"`
}

type InspectNode struct {
	Name       string    `json:"name"`
	TotalGPU   int       `json:"totalGPU"`
	UsedGPU    int       `json:"usedGPU"`
	Topology [][]TopologyType    `json:"topology"`
}