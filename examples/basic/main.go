package main

import (
	"fmt"
	"time"

	"github.com/jmeyers35/tempest/pkg/simulator"
)

func main() {
	// Use new configuration system for better reproducibility
	config := simulator.DefaultConfig()
	config.Seed = 12141997
	config.TickIncrement = 1 * time.Minute
	config.MaxTicks = 100 // Limit for demo
	config.FailureProbability = 0.1 // 10% chance of random failures
	config.EnableEventJitter = true // Add some timing variation
	
	sim := simulator.NewWithConfig(config)
	fmt.Printf("Simulator initialized with seed: %d\n", sim.GetSeed())

	wf := &Workflow{}

	sim.RegisterWorkflow(wf.DoALongBankTransaction)
	sim.RegisterInvariant(wf.DoALongBankTransaction, simulator.Invariant{
		Name: "balance must always be positive",
		Check: func() bool {
			return wf.Balance >= 0
		},
	})
	
	fmt.Println("Starting simulation...")
	err := sim.Run(wf.DoALongBankTransaction)
	if err != nil {
		fmt.Printf("Simulation error: %v\n", err)
	}
	
	// Print simulation statistics
	fmt.Printf("Random calls made: %d\n", sim.GetRandomCallCount())
	fmt.Printf("Events processed: %d\n", len(sim.GetEventHistory()))
}
