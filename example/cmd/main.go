package main

import (
	"fmt"
	"time"

	"github.com/jmeyers35/tempest/example"
	"github.com/jmeyers35/tempest/pkg/simulator"
)

func main() {
	sim := simulator.New(12141997)
	sim.SetTickDuration(1 * time.Minute)

	sim.RegisterWorkflow(example.Workflow)
	fmt.Println("starting simulator")
	sim.Run(example.Workflow)
}
