package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/scheduler"
)

func main() {

	nodeName := ""
	details := flag.Bool("d", false, "details")
	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		nodeName = args[0]
	}
	
	inspect, err := fetchNode(nodeName, *details)
	if err != nil {
		fmt.Printf("Failed due to %v", err)
		os.Exit(1)
	}
	fmt.Println(inspect)

	// todo display result

}

func fetchNode(node string, detail bool) (*scheduler.Result, error) {
	url := "http://127.0.0.1:32743/gputopology-scheduler/inspect"
	if len(node) != 0 {
		url += "/" + node
	}
	if detail {
		url += "?detail=true"
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexepected status code %d", resp.StatusCode)
	}
	rawData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var inspectResult scheduler.Result
	err = json.Unmarshal(rawData, &inspectResult)
	if err != nil {
		return nil, err
	}

	log.Printf("log: fetch node %v to inspect node info %v", node, inspectResult)

	return &inspectResult, nil
}
