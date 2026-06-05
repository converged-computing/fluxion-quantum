package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/converged-computing/fluxion-quantum/pkg/graph"
)

func main() {
	fmt.Println("This is the fluxion quantum resource matcher")
	confFilePath := flag.String("conf", "conf/quantum.json", "quantum.json that describes the graph structure")
	specFilePath := flag.String("spec", "", "JobSpec (yaml file) that defines ice cream request")
	matchPolicy := flag.String("policy", "first", "Match policy (defaults to first)")
	satisfy := flag.Bool("satisfy", false, "Request that the spec be assessed for satisfiability")
	flag.Parse()

	specFile := *specFilePath
	confFile := *confFilePath

	// The JobSpec file is required
	if specFile == "" {
		flag.Usage()
		os.Exit(0)
	}

	// The JobSpec file and graphml must exist
	if _, err := os.Stat(specFile); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("%s does not exist\n", specFile)
		os.Exit(0)
	}
	if _, err := os.Stat(confFile); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("%s does not exist\n", confFile)
		os.Exit(0)
	}

	// Create an ice cream graph, and match the spec to it.
	g := graph.FluxionGraph{}
	g.Init(confFile, *matchPolicy, "")

	if *satisfy {
		satisfied, err := g.Satisfy(specFile)
		if err != nil {
			fmt.Printf("The request cannot be satisfied: %s\n", err)
			return
		}
		if satisfied {
			fmt.Printf("\n😋 Your resources are satisfied.\n")
		} else {
			fmt.Printf("\n😭️ We could not satisfy your request.\n")
		}

	} else {
		allocation, err := g.MatchAllocate(specFile)
		if err != nil {
			fmt.Printf("The request cannot be satisfied: %s\n", err)
			return
		}
		allocation.Show()

	}
}
