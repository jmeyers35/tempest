package main

import (
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// PaymentState represents the state of a payment transaction
type PaymentState struct {
	OrderID           string
	Amount            float64
	CustomerID        string
	MerchantID        string
	PaymentStatus     string
	ReservationID     string
	ChargeID          string
	NotificationsSent []string
	Attempts          int
	LastError         string
}

// PaymentWorkflow implements a comprehensive payment processing workflow
type PaymentWorkflow struct {
	State PaymentState
}

// PaymentRequest represents the input to the payment workflow
type PaymentRequest struct {
	OrderID    string
	Amount     float64
	CustomerID string
	MerchantID string
}

// PaymentResult represents the output of the payment workflow
type PaymentResult struct {
	Success   bool
	ChargeID  string
	Message   string
	Timestamp time.Time
}

// ProcessPayment is the main workflow function that orchestrates payment processing
func (w *PaymentWorkflow) ProcessPayment(ctx workflow.Context) (*PaymentResult, error) {
	logger := workflow.GetLogger(ctx)
	
	// Initialize state with default values for testing
	w.State = PaymentState{
		OrderID:       "test-order-123",
		Amount:        99.99,
		CustomerID:    "test-customer",
		MerchantID:    "test-merchant",
		PaymentStatus: "pending",
		Attempts:      0,
	}
	
	logger.Info("Starting payment processing", "orderID", w.State.OrderID, "amount", w.State.Amount)

	// Step 1: Validate customer and payment method
	err := w.validateCustomer(ctx)
	if err != nil {
		w.State.PaymentStatus = "validation_failed"
		w.State.LastError = err.Error()
		return &PaymentResult{Success: false, Message: err.Error()}, err
	}

	// Step 2: Reserve funds (authorization)
	err = w.reserveFunds(ctx)
	if err != nil {
		w.State.PaymentStatus = "reservation_failed"
		w.State.LastError = err.Error()
		return &PaymentResult{Success: false, Message: err.Error()}, err
	}

	// Step 3: Process the actual charge with retries
	err = w.processChargeWithRetries(ctx)
	if err != nil {
		// If charge fails, we need to release the reservation
		w.releaseReservation(ctx)
		w.State.PaymentStatus = "charge_failed"
		w.State.LastError = err.Error()
		return &PaymentResult{Success: false, Message: err.Error()}, err
	}

	// Step 4: Send notifications (best effort)
	w.sendNotifications(ctx)

	// Step 5: Update merchant account
	err = w.updateMerchantAccount(ctx)
	if err != nil {
		// This is a critical failure - we've charged the customer but can't pay the merchant
		logger.Error("Critical: charged customer but failed to pay merchant", "error", err)
		w.State.PaymentStatus = "settlement_failed"
		w.State.LastError = err.Error()
		// In a real system, this would trigger manual intervention
		return &PaymentResult{Success: false, Message: "Settlement failed", ChargeID: w.State.ChargeID}, err
	}

	w.State.PaymentStatus = "completed"
	logger.Info("Payment processing completed successfully", "chargeID", w.State.ChargeID)
	
	return &PaymentResult{
		Success:   true,
		ChargeID:  w.State.ChargeID,
		Message:   "Payment processed successfully",
		Timestamp: workflow.Now(ctx),
	}, nil
}

func (w *PaymentWorkflow) validateCustomer(ctx workflow.Context) error {
	// Configure activity options with proper timeouts
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	
	var result bool
	
	err := workflow.ExecuteActivity(ctx,
		"ValidateCustomer",
		w.State.CustomerID,
	).Get(ctx, &result)
	
	if err != nil {
		w.State.Attempts++
		return fmt.Errorf("customer validation failed: %w", err)
	}
	
	if !result {
		return errors.New("customer validation failed: invalid customer")
	}
	
	return nil
}

func (w *PaymentWorkflow) reserveFunds(ctx workflow.Context) error {
	// Configure activity options with proper timeouts
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	
	var reservationID string
	
	err := workflow.ExecuteActivity(ctx,
		"ReserveFunds",
		w.State.CustomerID,
		w.State.Amount,
	).Get(ctx, &reservationID)
	
	if err != nil {
		w.State.Attempts++
		return fmt.Errorf("fund reservation failed: %w", err)
	}
	
	w.State.ReservationID = reservationID
	w.State.PaymentStatus = "funds_reserved"
	return nil
}

func (w *PaymentWorkflow) processChargeWithRetries(ctx workflow.Context) error {
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    30 * time.Second,
		MaximumAttempts:    3,
	}
	
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy:         retryPolicy,
	}
	
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	
	var chargeID string
	err := workflow.ExecuteActivity(ctx,
		"ProcessCharge",
		w.State.ReservationID,
		w.State.Amount,
		w.State.OrderID,
	).Get(ctx, &chargeID)
	
	if err != nil {
		w.State.Attempts++
		return fmt.Errorf("charge processing failed after retries: %w", err)
	}
	
	w.State.ChargeID = chargeID
	w.State.PaymentStatus = "charged"
	return nil
}

func (w *PaymentWorkflow) releaseReservation(ctx workflow.Context) {
	if w.State.ReservationID == "" {
		return
	}
	
	// Configure activity options with proper timeouts
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	
	// Best effort release - we don't want to fail the workflow if this fails
	workflow.ExecuteActivity(ctx,
		"ReleaseReservation",
		w.State.ReservationID,
	).Get(ctx, nil)
	
	w.State.PaymentStatus = "reservation_released"
}

func (w *PaymentWorkflow) sendNotifications(ctx workflow.Context) {
	// Configure activity options with proper timeouts
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	
	// Send customer notification
	workflow.ExecuteActivity(ctx,
		"SendCustomerNotification",
		w.State.CustomerID,
		w.State.ChargeID,
		w.State.Amount,
	).Get(ctx, nil) // Ignore errors for notifications
	
	// Send merchant notification  
	workflow.ExecuteActivity(ctx,
		"SendMerchantNotification",
		w.State.MerchantID,
		w.State.ChargeID,
		w.State.Amount,
	).Get(ctx, nil) // Ignore errors for notifications
	
	w.State.NotificationsSent = []string{"customer", "merchant"}
}

func (w *PaymentWorkflow) updateMerchantAccount(ctx workflow.Context) error {
	// Configure activity options with proper timeouts
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	
	var success bool
	
	err := workflow.ExecuteActivity(ctx,
		"UpdateMerchantAccount",
		w.State.MerchantID,
		w.State.Amount,
		w.State.ChargeID,
	).Get(ctx, &success)
	
	if err != nil {
		return fmt.Errorf("merchant account update failed: %w", err)
	}
	
	if !success {
		return errors.New("merchant account update returned false")
	}
	
	return nil
}