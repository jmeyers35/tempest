package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// PaymentActivities contains all the activities used by the payment workflow
type PaymentActivities struct{}

// ValidateCustomer checks if a customer is valid and has a good payment method
func (a *PaymentActivities) ValidateCustomer(ctx context.Context, customerID string) (bool, error) {
	// Simulate validation logic
	if customerID == "" {
		return false, errors.New("empty customer ID")
	}

	// Simulate some customers being invalid
	if customerID == "invalid-customer" {
		return false, nil
	}

	// Simulate validation taking some time
	// time.Sleep(100 * time.Millisecond) // Removed for testing

	return true, nil
}

// ReserveFunds reserves the specified amount for the customer
func (a *PaymentActivities) ReserveFunds(ctx context.Context, customerID string, amount float64) (string, error) {
	if amount <= 0 {
		return "", errors.New("invalid amount")
	}

	if amount > 10000 {
		return "", errors.New("amount exceeds daily limit")
	}

	// Simulate insufficient funds for some scenarios
	if customerID == "broke-customer" {
		return "", errors.New("insufficient funds")
	}

	// Generate a reservation ID
	reservationID := fmt.Sprintf("res_%d_%s", time.Now().Unix(), customerID)

	// Simulate processing time
	// time.Sleep(200 * time.Millisecond) // Removed for testing

	return reservationID, nil
}

// ProcessCharge processes the actual charge against the reserved funds
func (a *PaymentActivities) ProcessCharge(ctx context.Context, reservationID string, amount float64, orderID string) (string, error) {
	if reservationID == "" {
		return "", errors.New("empty reservation ID")
	}

	// Simulate payment processor being down occasionally
	if reservationID == "fail-reservation" {
		return "", errors.New("payment processor unavailable")
	}

	// Generate a charge ID
	chargeID := fmt.Sprintf("chg_%d_%s", time.Now().Unix(), orderID)

	// Simulate charge processing time
	// time.Sleep(300 * time.Millisecond) // Removed for testing

	return chargeID, nil
}

// ReleaseReservation releases a fund reservation (compensation activity)
func (a *PaymentActivities) ReleaseReservation(ctx context.Context, reservationID string) error {
	if reservationID == "" {
		return errors.New("empty reservation ID")
	}

	// Simulate release processing
	// time.Sleep(100 * time.Millisecond) // Removed for testing

	return nil
}

// SendCustomerNotification sends a notification to the customer
func (a *PaymentActivities) SendCustomerNotification(ctx context.Context, customerID string, chargeID string, amount float64) error {
	// Simulate notification service being unreliable
	if rand.Float64() < 0.1 { // 10% failure rate
		return errors.New("notification service unavailable")
	}

	// Simulate sending notification
	// time.Sleep(50 * time.Millisecond) // Removed for testing

	return nil
}

// SendMerchantNotification sends a notification to the merchant
func (a *PaymentActivities) SendMerchantNotification(ctx context.Context, merchantID string, chargeID string, amount float64) error {
	// Simulate notification service being unreliable
	if rand.Float64() < 0.15 { // 15% failure rate
		return errors.New("merchant notification service unavailable")
	}

	// Simulate sending notification
	// time.Sleep(75 * time.Millisecond) // Removed for testing

	return nil
}

// UpdateMerchantAccount updates the merchant's account with the payment
func (a *PaymentActivities) UpdateMerchantAccount(ctx context.Context, merchantID string, amount float64, chargeID string) (bool, error) {
	if merchantID == "" {
		return false, errors.New("empty merchant ID")
	}

	// Simulate merchant account service being occasionally unavailable
	if merchantID == "suspended-merchant" {
		return false, errors.New("merchant account suspended")
	}

	// Simulate processing time
	// time.Sleep(250 * time.Millisecond) // Removed for testing

	return true, nil
}
