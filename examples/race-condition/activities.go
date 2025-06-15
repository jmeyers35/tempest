package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// InventorySystem represents a shared inventory state
// This is intentionally designed with a race condition
type InventorySystem struct {
	mu           sync.RWMutex
	inventory    map[string]int // productID -> available quantity
	reservations map[string]int // productID -> reserved quantity
}

var GlobalInventory = &InventorySystem{
	inventory: map[string]int{
		"laptop":   10,
		"mouse":    50,
		"keyboard": 25,
		"monitor":  15,
	},
	reservations: make(map[string]int),
}

// OrderActivities contains all the activities used by the order workflow
type OrderActivities struct{}

// CheckInventory checks if the requested quantity is available
// This activity has the race condition - it checks but doesn't reserve
func (a *OrderActivities) CheckInventory(ctx context.Context, productID string, quantity int) (bool, error) {
	if productID == "" || quantity <= 0 {
		return false, errors.New("invalid product ID or quantity")
	}

	GlobalInventory.mu.RLock()
	defer GlobalInventory.mu.RUnlock()

	available := GlobalInventory.inventory[productID]
	reserved := GlobalInventory.reservations[productID]

	// The bug: we check availability but don't atomically reserve
	// Another workflow could reserve items between this check and ReserveItems
	canFulfill := (available - reserved) >= quantity

	// Simulate some processing time that makes the race condition more likely
	// In a real system, this could be network calls, database queries, etc.
	// time.Sleep(10 * time.Millisecond) // Commented out for deterministic testing

	return canFulfill, nil
}

// ReserveItems reserves the specified quantity (this is where the race happens)
func (a *OrderActivities) ReserveItems(ctx context.Context, productID string, quantity int, orderID string) (string, error) {
	if productID == "" || quantity <= 0 || orderID == "" {
		return "", errors.New("invalid parameters")
	}

	GlobalInventory.mu.Lock()
	defer GlobalInventory.mu.Unlock()

	reserved := GlobalInventory.reservations[productID]

	// BUG: We don't re-check availability here!
	// We assume that because CheckInventory passed, we can still reserve
	// But another workflow might have reserved items in the meantime

	GlobalInventory.reservations[productID] = reserved + quantity

	reservationID := fmt.Sprintf("res_%s_%s_%d", orderID, productID, quantity)
	return reservationID, nil
}

// ProcessPayment simulates payment processing
func (a *OrderActivities) ProcessPayment(ctx context.Context, orderID string, amount float64) (string, error) {
	if orderID == "" || amount <= 0 {
		return "", errors.New("invalid order ID or amount")
	}

	// Simulate payment processing
	paymentID := fmt.Sprintf("pay_%s_%.2f", orderID, amount)
	return paymentID, nil
}

// FulfillOrder moves reserved items to fulfilled (removes from inventory)
func (a *OrderActivities) FulfillOrder(ctx context.Context, productID string, quantity int, reservationID string) error {
	if productID == "" || quantity <= 0 || reservationID == "" {
		return errors.New("invalid parameters")
	}

	GlobalInventory.mu.Lock()
	defer GlobalInventory.mu.Unlock()

	// Move from reserved to fulfilled (reduce actual inventory)
	GlobalInventory.inventory[productID] -= quantity
	GlobalInventory.reservations[productID] -= quantity

	return nil
}

// CancelReservation cancels a reservation (compensation activity)
func (a *OrderActivities) CancelReservation(ctx context.Context, productID string, quantity int, reservationID string) error {
	if productID == "" || quantity <= 0 {
		return errors.New("invalid parameters")
	}

	GlobalInventory.mu.Lock()
	defer GlobalInventory.mu.Unlock()

	// Remove reservation
	GlobalInventory.reservations[productID] -= quantity
	if GlobalInventory.reservations[productID] < 0 {
		GlobalInventory.reservations[productID] = 0
	}

	return nil
}

// GetInventoryStatus returns current inventory levels (for debugging/invariants)
func (GlobalInventory *InventorySystem) GetStatus() map[string]struct {
	Available int
	Reserved  int
	Total     int
} {
	GlobalInventory.mu.RLock()
	defer GlobalInventory.mu.RUnlock()

	status := make(map[string]struct {
		Available int
		Reserved  int
		Total     int
	})

	for productID, total := range GlobalInventory.inventory {
		reserved := GlobalInventory.reservations[productID]
		status[productID] = struct {
			Available int
			Reserved  int
			Total     int
		}{
			Available: total - reserved,
			Reserved:  reserved,
			Total:     total,
		}
	}

	return status
}
