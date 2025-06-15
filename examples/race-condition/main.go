package main

import (
	"fmt"
	"time"

	"github.com/jmeyers35/tempest/pkg/simulator"
)

func main() {
	fmt.Println("🏪 Race Condition Detection Demo")
	fmt.Println("Testing concurrent inventory management workflows")

	// Configure simulation for race condition detection
	config := simulator.DefaultConfig()
	config.Seed = 42 // Fixed seed for reproducible race conditions
	config.TickIncrement = 10 * time.Millisecond // Small ticks to increase concurrency
	config.MaxTicks = 1000
	config.EnableEventJitter = true // Add timing variability

	sim := simulator.NewWithConfig(config)
	fmt.Printf("🎯 Simulation seed: %d\n", sim.GetSeed())

	// Reset global inventory state for each run
	GlobalInventory = &InventorySystem{
		inventory: map[string]int{
			"laptop":   10,
			"mouse":    50,
			"keyboard": 25,
			"monitor":  15,
		},
		reservations: make(map[string]int),
	}

	// Create activities
	activities := &OrderActivities{}

	// Register activities - we want them to actually execute their real logic
	// to interact with the shared GlobalInventory state
	sim.RegisterActivity(activities.CheckInventory)
	sim.RegisterActivity(activities.ReserveItems) 
	sim.RegisterActivity(activities.ProcessPayment)
	sim.RegisterActivity(activities.FulfillOrder)
	sim.RegisterActivity(activities.CancelReservation)

	// Create order batch processor workflow
	batchProcessor := &OrderBatchProcessor{}
	
	// Define the orders to process concurrently
	orders := []OrderRequest{
		{OrderID: "order-001", ProductID: "laptop", Quantity: 3, Amount: 2997.00, CustomerID: "cust-001"},
		{OrderID: "order-002", ProductID: "laptop", Quantity: 4, Amount: 3996.00, CustomerID: "cust-002"},
		{OrderID: "order-003", ProductID: "laptop", Quantity: 5, Amount: 4995.00, CustomerID: "cust-003"},
		{OrderID: "order-004", ProductID: "laptop", Quantity: 2, Amount: 1998.00, CustomerID: "cust-004"},
		{OrderID: "order-005", ProductID: "laptop", Quantity: 3, Amount: 2997.00, CustomerID: "cust-005"},
	}

	// Register the batch workflow
	sim.RegisterWorkflow(batchProcessor.ProcessBatchOrders)

	// Add critical invariant: Total reservations should never exceed available inventory
	sim.RegisterInvariant(batchProcessor.ProcessBatchOrders, simulator.Invariant{
		Name: "no overselling - reservations cannot exceed inventory",
		Check: func() bool {
			status := GlobalInventory.GetStatus()
			for productID, info := range status {
				if info.Reserved > info.Total {
					fmt.Printf("🚨 INVARIANT VIOLATION: Product %s has %d reserved but only %d total inventory!\n",
						productID, info.Reserved, info.Total)
					return false
				}
				// Also check that available inventory is not negative
				if info.Available < 0 {
					fmt.Printf("🚨 INVARIANT VIOLATION: Product %s has negative available inventory: %d\n",
						productID, info.Available)
					return false
				}
			}
			return true
		},
	})

	// Add another invariant: Reserved + Available should equal Total
	sim.RegisterInvariant(batchProcessor.ProcessBatchOrders, simulator.Invariant{
		Name: "inventory accounting consistency",
		Check: func() bool {
			status := GlobalInventory.GetStatus()
			for productID, info := range status {
				if info.Reserved+info.Available != info.Total {
					fmt.Printf("🚨 INVARIANT VIOLATION: Product %s accounting mismatch: %d reserved + %d available != %d total\n",
						productID, info.Reserved, info.Available, info.Total)
					return false
				}
			}
			return true
		},
	})

	fmt.Println("\n📦 Initial Inventory Status:")
	status := GlobalInventory.GetStatus()
	for productID, info := range status {
		fmt.Printf("  %s: %d total, %d available, %d reserved\n",
			productID, info.Total, info.Available, info.Reserved)
	}

	fmt.Println("\n🚀 Starting concurrent order processing simulation...")
	fmt.Println("Note: Total laptop orders = 17, but only 10 laptops available")
	fmt.Println("This should trigger race condition and invariant violations!")

	// Run the batch workflow that processes all orders concurrently
	// Using the new enhanced API that accepts input parameters
	err := sim.Run(batchProcessor.ProcessBatchOrders, orders)
	if err != nil {
		fmt.Printf("Batch workflow failed: %v\n", err)
	}

	// Check inventory during processing to see what happened
	fmt.Println("\n🔍 Checking what actually happened to inventory...")
	midStatus := GlobalInventory.GetStatus()
	for productID, info := range midStatus {
		fmt.Printf("  %s: %d total, %d available, %d reserved\n",
			productID, info.Total, info.Available, info.Reserved)
	}

	// Results
	fmt.Println("\n📊 Final Results:")
	finalStatus := GlobalInventory.GetStatus()
	for productID, info := range finalStatus {
		fmt.Printf("  %s: %d total, %d available, %d reserved\n",
			productID, info.Total, info.Available, info.Reserved)
	}

	// Show workflow states
	fmt.Println("\n📋 Order Status:")
	for i, state := range batchProcessor.OrderStates {
		fmt.Printf("  Order %d (%s): %s", i+1, state.OrderID, state.Status)
		if state.ErrorMsg != "" {
			fmt.Printf(" - %s", state.ErrorMsg)
		}
		fmt.Println()
	}

	fmt.Printf("\nRandom calls made: %d\n", sim.GetRandomCallCount())

	fmt.Println("\n✅ Demo completed!")
	fmt.Println("\n🎯 Race Condition Analysis:")
	fmt.Println("If you see an invariant violation panic above, the race condition was successfully detected!")
	fmt.Println()
	fmt.Println("🐛 The Bug Explained:")
	fmt.Println("1. All 5 orders checked inventory individually (each saw 10 laptops available)")
	fmt.Println("2. All 5 orders passed the availability check")  
	fmt.Println("3. All 5 orders then reserved items: 3+4+5+2+3 = 17 laptops total")
	fmt.Println("4. But only 10 laptops were actually available!")
	fmt.Println("5. Result: Inventory went negative, triggering invariant violation")
	fmt.Println()
	fmt.Println("🔍 Why Traditional Tests Miss This:")
	fmt.Println("- Unit tests: Test activities in isolation")
	fmt.Println("- Integration tests: Usually run workflows sequentially") 
	fmt.Println("- Temporal test server: Activities often mocked or run synchronously")
	fmt.Println()
	fmt.Println("⚡ How Tempest Caught It:")
	fmt.Println("- Deterministic concurrent execution of the workflow")
	fmt.Println("- Real activity implementations with shared state")
	fmt.Println("- Invariant checking after each simulation tick")
	fmt.Println("- Reproducible results with fixed seeds")
}