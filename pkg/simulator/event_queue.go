package simulator

import (
	"sync"
)

// EventQueue is a priority queue for deterministic event scheduling
type EventQueue struct {
	events []*Event
	mu     sync.Mutex
}

// Push adds an event to the queue, maintaining time and priority order
func (eq *EventQueue) Push(event *Event) {
	eq.mu.Lock()
	defer eq.mu.Unlock()
	eq.events = append(eq.events, event)
	// Sort by time, then by priority
	for i := len(eq.events) - 1; i > 0; i-- {
		if eq.events[i].Time.Before(eq.events[i-1].Time) ||
			(eq.events[i].Time.Equal(eq.events[i-1].Time) && eq.events[i].Priority < eq.events[i-1].Priority) {
			eq.events[i], eq.events[i-1] = eq.events[i-1], eq.events[i]
		} else {
			break
		}
	}
}

// Pop removes and returns the highest priority event from the queue
func (eq *EventQueue) Pop() *Event {
	eq.mu.Lock()
	defer eq.mu.Unlock()
	if len(eq.events) == 0 {
		return nil
	}
	event := eq.events[0]
	eq.events = eq.events[1:]
	return event
}

// Peek returns the highest priority event without removing it
func (eq *EventQueue) Peek() *Event {
	eq.mu.Lock()
	defer eq.mu.Unlock()
	if len(eq.events) == 0 {
		return nil
	}
	return eq.events[0]
}

// Size returns the number of events in the queue
func (eq *EventQueue) Size() int {
	eq.mu.Lock()
	defer eq.mu.Unlock()
	return len(eq.events)
}