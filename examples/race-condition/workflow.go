package main

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

// OrderRequest represents an order to be processed
type OrderRequest struct {
	OrderID    string
	ProductID  string
	Quantity   int
	Amount     float64
	CustomerID string
}

// OrderState tracks the state of an order workflow
type OrderState struct {
	OrderID       string
	Status        string // "pending", "reserved", "paid", "fulfilled", "failed", "cancelled"
	ProductID     string
	Quantity      int
	ReservationID string
	PaymentID     string
	ErrorMsg      string
}

// OrderWorkflow processes customer orders with a race condition bug
type OrderWorkflow struct {
	State OrderState
}

// OrderBatchProcessor handles concurrent order processing
type OrderBatchProcessor struct {
	OrderStates []OrderState
}

// ProcessOrder is the main workflow that contains the race condition
func (w *OrderWorkflow) ProcessOrder(ctx workflow.Context, request OrderRequest) error {
	// Initialize state
	w.State = OrderState{
		OrderID:   request.OrderID,
		Status:    "pending",
		ProductID: request.ProductID,
		Quantity:  request.Quantity,
	}
	
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting order processing", "orderID", request.OrderID, "productID", request.ProductID, "quantity", request.Quantity)

	// Activity options with retries
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 1 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Step 1: Check inventory availability
	var isAvailable bool
	err := workflow.ExecuteActivity(ctx, "CheckInventory", request.ProductID, request.Quantity).Get(ctx, &isAvailable)
	if err != nil {
		w.State.Status = "failed"
		w.State.ErrorMsg = fmt.Sprintf("inventory check failed: %v", err)
		return err
	}

	if !isAvailable {
		w.State.Status = "failed"
		w.State.ErrorMsg = "insufficient inventory"
		logger.Info("Order failed - insufficient inventory")
		return fmt.Errorf("insufficient inventory for product %s", request.ProductID)
	}

	// RACE CONDITION BUG: Between the inventory check above and the reservation below,
	// another workflow might have reserved the same items!
	// The gap here is where the race condition occurs.
	
	// In a real system, there might be additional processing here:
	// - Validate customer information
	// - Check pricing
	// - Apply discounts
	// - etc.
	// This processing time makes the race condition more likely to occur.

	// Step 2: Reserve the items (this is where overselling can happen)
	var reservationID string
	err = workflow.ExecuteActivity(ctx, "ReserveItems", request.ProductID, request.Quantity, request.OrderID).Get(ctx, &reservationID)
	if err != nil {
		w.State.Status = "failed"
		w.State.ErrorMsg = fmt.Sprintf("reservation failed: %v", err)
		return err
	}

	w.State.Status = "reserved"
	w.State.ReservationID = reservationID
	logger.Info("Items reserved successfully", "reservationID", reservationID)

	// Step 3: Process payment
	var paymentID string
	err = workflow.ExecuteActivity(ctx, "ProcessPayment", request.OrderID, request.Amount).Get(ctx, &paymentID)
	if err != nil {
		// Compensation: Cancel the reservation
		compensationErr := workflow.ExecuteActivity(ctx, "CancelReservation", request.ProductID, request.Quantity, reservationID).Get(ctx, nil)
		if compensationErr != nil {
			logger.Error("Failed to cancel reservation during compensation", "error", compensationErr)
		}
		
		w.State.Status = "failed"
		w.State.ErrorMsg = fmt.Sprintf("payment failed: %v", err)
		return err
	}

	w.State.Status = "paid"
	w.State.PaymentID = paymentID
	logger.Info("Payment processed successfully", "paymentID", paymentID)

	// Step 4: Fulfill the order
	err = workflow.ExecuteActivity(ctx, "FulfillOrder", request.ProductID, request.Quantity, reservationID).Get(ctx, nil)
	if err != nil {
		// In a real system, we'd need more sophisticated compensation logic here
		logger.Error("Failed to fulfill order", "error", err)
		w.State.Status = "failed"
		w.State.ErrorMsg = fmt.Sprintf("fulfillment failed: %v", err)
		return err
	}

	w.State.Status = "fulfilled"
	logger.Info("Order fulfilled successfully")
	
	return nil
}

// ProcessBatchOrders processes multiple orders concurrently to expose race conditions
// This is now an idiomatic Temporal workflow that accepts input parameters
func (w *OrderBatchProcessor) ProcessBatchOrders(ctx workflow.Context, orders []OrderRequest) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting batch order processing", "orderCount", len(orders))

	// Set activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 1 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Initialize states for all orders
	w.OrderStates = make([]OrderState, len(orders))
	for i, order := range orders {
		w.OrderStates[i] = OrderState{
			OrderID:   order.OrderID,
			Status:    "pending",
			ProductID: order.ProductID,
			Quantity:  order.Quantity,
		}
	}

	// Process all orders - the race condition happens because we check inventory
	// for all orders first, then try to reserve for all of them
	
	// Step 1: Check inventory for all orders (all will pass)
	inventoryChecks := make([]bool, len(orders))
	for i, order := range orders {
		var isAvailable bool
		err := workflow.ExecuteActivity(ctx, "CheckInventory", order.ProductID, order.Quantity).Get(ctx, &isAvailable)
		if err != nil {
			logger.Error("Inventory check failed", "orderID", order.OrderID, "error", err)
			w.OrderStates[i].Status = "failed"
			w.OrderStates[i].ErrorMsg = fmt.Sprintf("inventory check failed: %v", err)
			continue
		}
		inventoryChecks[i] = isAvailable
		if !isAvailable {
			w.OrderStates[i].Status = "failed"
			w.OrderStates[i].ErrorMsg = "insufficient inventory"
		}
	}

	// Step 2: Reserve items for all orders that passed inventory check
	// THIS IS THE BUG: All orders passed check, but total demand exceeds supply
	for i, order := range orders {
		if !inventoryChecks[i] {
			continue // Skip orders that failed inventory check
		}
		
		var reservationID string
		err := workflow.ExecuteActivity(ctx, "ReserveItems", order.ProductID, order.Quantity, order.OrderID).Get(ctx, &reservationID)
		if err != nil {
			logger.Error("Reservation failed", "orderID", order.OrderID, "error", err)
			w.OrderStates[i].Status = "failed"
			w.OrderStates[i].ErrorMsg = fmt.Sprintf("reservation failed: %v", err)
			continue
		}

		w.OrderStates[i].Status = "reserved"
		w.OrderStates[i].ReservationID = reservationID
		logger.Info("Items reserved", "orderID", order.OrderID, "reservationID", reservationID)

		// Continue with payment and fulfillment for successful reservations
		var paymentID string
		err = workflow.ExecuteActivity(ctx, "ProcessPayment", order.OrderID, order.Amount).Get(ctx, &paymentID)
		if err != nil {
			// Compensation: Cancel the reservation
			workflow.ExecuteActivity(ctx, "CancelReservation", order.ProductID, order.Quantity, reservationID).Get(ctx, nil)
			w.OrderStates[i].Status = "failed"
			w.OrderStates[i].ErrorMsg = fmt.Sprintf("payment failed: %v", err)
			continue
		}

		w.OrderStates[i].Status = "paid"
		w.OrderStates[i].PaymentID = paymentID

		// Fulfill the order
		err = workflow.ExecuteActivity(ctx, "FulfillOrder", order.ProductID, order.Quantity, reservationID).Get(ctx, nil)
		if err != nil {
			w.OrderStates[i].Status = "failed"
			w.OrderStates[i].ErrorMsg = fmt.Sprintf("fulfillment failed: %v", err)
			continue
		}

		w.OrderStates[i].Status = "fulfilled"
		logger.Info("Order fulfilled", "orderID", order.OrderID)
	}

	// Log final status
	for _, state := range w.OrderStates {
		logger.Info("Order completed", "orderID", state.OrderID, "status", state.Status)
	}

	return nil
}

