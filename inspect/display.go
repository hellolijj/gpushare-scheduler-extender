package main

import (
	"bytes"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/types"
)

func display(nodes []*types.InspectNode) {
	
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	
	// print title
	var buffer bytes.Buffer
	buffer.WriteString("NAME\tALLGPU\tUSEDGPU\n")
	
	// print content
	for _, node := range nodes {
		buffer.WriteString(fmt.Sprintf("%s\t%d\t%d\n", node.Name, node.TotalGPU, node.UsedGPU))
	}
	fmt.Fprintf(w, buffer.String())
	
	
	w.Flush()
}
