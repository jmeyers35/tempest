package main

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/jmeyers35/tempest/pkg/simulator"
)

func main() {
	fmt.Println("🏦 Payment Workflow DST Demo")
	fmt.Println("Testing activity-level fault injection")

	// Configure simulation
	config := simulator.DefaultConfig()
	config.Seed = 123456
	config.TickIncrement = 30 * time.Second
	config.MaxTicks = 200
	config.EnableEventJitter = true

	sim, err := simulator.NewWithConfig(config)
	if err != nil {
		fmt.Printf("Failed to create simulator: %v\n", err)
		return
	}
	fmt.Printf("🎯 Simulation seed: %d\n", sim.GetSeed())

	// Create activities
	activities := &PaymentActivities{}

	// Register activities with realistic failure behaviors
	sim.RegisterActivityWithBehavior(activities.ValidateCustomer, simulator.ActivityBehavior{
		Name:        "ValidateCustomer",
		FailureRate: 0.1, // 10% validation failures
		MinLatency:  50 * time.Millisecond,
		MaxLatency:  200 * time.Millisecond,
		ReturnValue: true,
		CustomErrors: []error{
			errors.New("customer service unavailable"),
			errors.New("invalid customer credentials"),
		},
	})

	sim.RegisterActivityWithBehavior(activities.ProcessCharge, simulator.ActivityBehavior{
		Name:        "ProcessCharge",
		FailureRate: 0.15, // 15% charge failures
		TimeoutRate: 0.05, // 5% timeouts
		MinLatency:  200 * time.Millisecond,
		MaxLatency:  1 * time.Second,
		ReturnValue: "charge_12345",
		CustomErrors: []error{
			errors.New("payment processor unavailable"),
			errors.New("card declined"),
		},
	})

	// Register all other activities needed by the workflow
	sim.RegisterActivityWithBehavior(activities.ReserveFunds, simulator.ActivityBehavior{
		Name:        "ReserveFunds",
		FailureRate: 0.05,
		ReturnValue: "reservation_12345",
	})

	sim.RegisterActivityWithBehavior(activities.ReleaseReservation, simulator.ActivityBehavior{
		Name:        "ReleaseReservation",
		FailureRate: 0.01,
		ErrorOnly:   true,
	})

	sim.RegisterActivityWithBehavior(activities.SendCustomerNotification, simulator.ActivityBehavior{
		Name:        "SendCustomerNotification",
		FailureRate: 0.1,
		ErrorOnly:   true,
	})

	sim.RegisterActivityWithBehavior(activities.SendMerchantNotification, simulator.ActivityBehavior{
		Name:        "SendMerchantNotification",
		FailureRate: 0.15,
		ErrorOnly:   true,
	})

	sim.RegisterActivityWithBehavior(activities.UpdateMerchantAccount, simulator.ActivityBehavior{
		Name:        "UpdateMerchantAccount",
		FailureRate: 0.05,
		ReturnValue: true,
	})

	// Create a simple test workflow that uses activities
	workflow := &PaymentWorkflow{}
	sim.RegisterWorkflow(workflow.ProcessPayment)

	// Add business invariants
	sim.RegisterInvariant(workflow.ProcessPayment, simulator.Invariant{
		Name: "payment status is valid",
		Check: func() bool {
			validStatuses := []string{"pending", "completed", "failed", "validation_failed", "charge_failed"}
			if slices.Contains(validStatuses, workflow.State.PaymentStatus) {
				return true
			}
			return workflow.State.PaymentStatus == "" // Initial state
		},
	})

	fmt.Println("\n🚀 Starting simulation...")
	err = sim.Run(workflow.ProcessPayment)

	// Results
	fmt.Println("\n📊 Results:")
	fmt.Printf("  Error: %v\n", err)
	fmt.Printf("  Activity calls made: %d\n", len(sim.GetActivityCalls()))
	fmt.Printf("  Faults injected: %d\n", len(sim.GetFaultHistory()))
	fmt.Printf("  Random calls: %d\n", sim.GetRandomCallCount())

	// Show activity call details
	if len(sim.GetActivityCalls()) > 0 {
		fmt.Println("\n🎬 Activity Calls:")
		for i, call := range sim.GetActivityCalls() {
			status := "✅ Success"
			if call.Error != nil {
				status = fmt.Sprintf("❌ Failed: %v", call.Error)
			}
			fmt.Printf("  %d. %s - %s (latency: %v)\n", i+1, call.Name, status, call.Latency)
		}
	}

	fmt.Println("\n✅ Demo completed!")
}
