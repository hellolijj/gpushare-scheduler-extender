package main

import (
	"fmt"
	"strings"
)

func main() {
	s := ""
	fmt.Print(len(strings.Split(s, ",")))


}

func test1()  {
	annotation := make(map[string]string)
	
	annotation["abced"] = "afag"
	annotation["abce"] = "ag"
	annotation["abc"] = "afg"
	annotation["GSOC_DEV_0_0"] = "0"
	annotation["GSOC_DEV_0_1"] = "1"
	annotation["GSOC_DEV_1_0"] = "1"
	annotation["GSOC_DEV_1_1"] = "0"
	
	topology := make(map[uint]map[uint]uint)
	
	for k, v := range annotation {
		if strings.HasPrefix(k, "GSOC_DEV_") {
			var gpu1, gpu2, topo uint
			fmt.Sscanf(k, "GSOC_DEV_%d_%d", &gpu1, &gpu2)
			fmt.Sscanf(v, "%d", &topo)
			topology[gpu1] = map[uint]uint{gpu2: topo}
		}
	}
	
}
