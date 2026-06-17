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
	retval := flag.Bool("retval", false, "Return value should reflect match success/failure")
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
		// Exit on not satisfied
		exitOnError(err, *retval)
		if satisfied {
			fmt.Printf("\n😋 Your resources are satisfied.\n")
		} else {
			fmt.Printf("\n😭️ We could not satisfy your request.\n")
		}

	} else {
		// Always an error if not satsified?
		allocation, err := g.MatchAllocate(specFile)
		exitOnError(err, *retval)
		allocation.Show()
		exitNotSatisfied(allocation.Satisfied(), *retval)
	}
}

// notSatisfied is a shared function to exit based on result and
// request for a return value
func exitNotSatisfied(satisfied, exitRetVal bool) {
	if !satisfied {
		if exitRetVal {
			os.Exit(1)
		}
	}
}

func exitOnError(err error, exitRetVal bool) {
	if err != nil {
		fmt.Printf("The request cannot be satisfied: %s\n", err)
		if exitRetVal {
			os.Exit(1)
		}
	}
}
