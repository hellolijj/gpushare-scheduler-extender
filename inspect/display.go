package main

import (
	"bytes"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/types"
)

const (
	legend = `
Legend:
 X    = Self
 SYS  = Connection traversing PCIe as well as the SMP interconnect between NUMA nodes (e.g., QPI/UPI)
 NODE = Connection traversing PCIe as well as the interconnect between PCIe Host Bridges within a NUMA node
 PHB  = Connection traversing PCIe as well as a PCIe Host Bridge (typically the CPU)
 PXB  = Connection traversing multiple PCIe switches (without traversing the PCIe Host Bridge)
 PIX  = Connection traversing a single PCIe switch
 PSB  = Connection traversing a single on-board PCIe switch
 NV#  = Connection traversing a bonded set of # NVLinks`
)


func displaySummary(nodes []*types.InspectNode) {
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

func displayDetails(res *types.InspectResult) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	
	var bufferPolicy bytes.Buffer
	bufferPolicy.WriteString(fmt.Sprintf("gpu policy: %s\n", res.Policy))
	fmt.Fprintf(w, bufferPolicy.String())
	fmt.Fprintln(w, "----------------------------------------")
	
	
	// print content
	for i, node := range res.Nodes {
		var bufferContent bytes.Buffer
		
		bufferContent.WriteString(fmt.Sprintf("node name: %s\n", node.Name))
		bufferContent.WriteString(fmt.Sprintf("all gpu: %d\n", node.TotalGPU))
		bufferContent.WriteString(fmt.Sprintf("used gpu: %d\n", node.UsedGPU))
		
		// print node type
		if len(node.NodeType) != 0 {
			bufferContent.WriteString(fmt.Sprintf("node type: %s\n", node.NodeType))
			// TODO: desplay node type details
		}
		
		// print topology
		if node.Topology != nil {
			bufferContent.WriteString(fmt.Sprintln("gpu topolocy"))
			numGpus := len(node.Topology)
			
			// print topology header line
			bufferContent.WriteString(fmt.Sprintf("     "))
			for i := 0; i < numGpus; i++ {
				bufferContent.WriteString(fmt.Sprintf(" GPU%d", i))
			}
			bufferContent.WriteString("\n")
			
			for i := 0; i < numGpus; i++ {
				bufferContent.WriteString(fmt.Sprintf("GPU%d ", i))
				for j := 0; j < numGpus; j++ {
					if node.Topology[i][j] == 0 {
						bufferContent.WriteString(fmt.Sprintf("%5s", "X"))
					} else {
						bufferContent.WriteString(fmt.Sprintf("%5d", node.Topology[i][j].Abbr()))
					}
					
				}
				bufferContent.WriteString("\n")
			}
		}
		
		
		if i < len(res.Nodes)-1 {
			bufferContent.WriteString("----------------------------------------\n")
		}
		fmt.Fprintf(w, bufferContent.String())
	}
	
	fmt.Fprintln(w, legend)
	
	w.Flush()
}
