package main

import (
	"fmt"
	"time"

	"github.com/jmeyers35/tempest/pkg/simulator"
)

func main() {
	// Configure the simulator with fault injection
	config := simulator.DefaultConfig()
	config.Seed = 42 // Fixed seed for reproducible testing
	config.TickIncrement = 1 * time.Minute
	config.MaxTicks = 200 // Allow more ticks for fault scenarios
	config.EnableEventJitter = true

	// Add fault injection configurations
	config.FaultConfigs = []simulator.FaultConfig{
		{
			Type:        simulator.FaultActivityError,
			Probability: 0.05, // 5% chance of activity failures
		},
		{
			Type:        simulator.FaultNetworkDelay,
			Probability: 0.1, // 10% chance of network delays
		},
		{
			Type:        simulator.FaultActivityTimeout,
			Probability: 0.03, // 3% chance of timeouts
		},
	}

	sim, err := simulator.NewWithConfig(config)
	if err != nil {
		fmt.Printf("Failed to create simulator: %v\n", err)
		return
	}
	fmt.Printf("🚀 Fault Injection Demo - Seed: %d\n", sim.GetSeed())

	// Create workflow for fault injection testing
	wf := &PaymentWorkflow{}

	sim.RegisterWorkflow(wf.ProcessPayment)

	// Add multiple invariants
	sim.RegisterInvariant(wf.ProcessPayment, simulator.Invariant{
		Name: "payment status is valid",
		Check: func() bool {
			validStatuses := []string{"pending", "completed", "failed", "validation_failed", "charge_failed"}
			for _, status := range validStatuses {
				if wf.State.PaymentStatus == status {
					return true
				}
			}
			return wf.State.PaymentStatus == "" // Initial state
		},
	})

	// Manually inject some faults for demonstration
	sim.InjectActivityFault("bank_api_call", fmt.Errorf("bank API temporarily unavailable"))
	sim.InjectNetworkDelay("database_connection", 2*time.Second)

	fmt.Println("🎯 Starting simulation with fault injection...")
	err = sim.Run(wf.ProcessPayment)

	// Print comprehensive results
	fmt.Println("\n📊 Simulation Results:")
	fmt.Printf("  Error: %v\n", err)
	fmt.Printf("  Payment Status: %s\n", wf.State.PaymentStatus)
	fmt.Printf("  Random calls made: %d\n", sim.GetRandomCallCount())
	fmt.Printf("  Events processed: %d\n", len(sim.GetEventHistory()))
	fmt.Printf("  Faults injected: %d\n", len(sim.GetFaultHistory()))

	// Print fault history
	if len(sim.GetFaultHistory()) > 0 {
		fmt.Println("\n💥 Fault History:")
		for i, fault := range sim.GetFaultHistory() {
			fmt.Printf("  %d. [Tick %d] %s on %s: %s\n",
				i+1, fault.Tick, fault.Type.String(), fault.Target, fault.Details)
		}
	}

	// Print recent events
	fmt.Println("\n📋 Recent Events:")
	events := sim.GetEventHistory()
	start := len(events) - 5
	if start < 0 {
		start = 0
	}
	for i := start; i < len(events); i++ {
		event := events[i]
		fmt.Printf("  [%d] %s: %s\n",
			event.Tick, event.Type.String(), event.Details)
	}
}
