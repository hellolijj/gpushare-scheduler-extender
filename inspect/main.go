package main

import (
	"flag"
	"fmt"
)

func main() {

	policy := flag.String("mode", "", "config gpu select policy, , for more detail: https://github.com/hellolijj/")
	staticdgx := flag.String("staticdgx", "", "config static node dgx, for more detail: https://github.com/hellolijj/")
	flag.Parse()
	fmt.Println(*policy)
	fmt.Println(*staticdgx)
	flag.CommandLine.Name()

}
