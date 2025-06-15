package simulator

import (
	"container/heap"
	"sync"
)

// EventQueue is a priority queue for deterministic event scheduling
type EventQueue struct {
	events eventHeap
	mu     sync.Mutex
}

// eventHeap implements heap.Interface for Event pointers
type eventHeap []*Event

func (h eventHeap) Len() int {
	return len(h)
}

func (h eventHeap) Less(i, j int) bool {
	// Sort by time first, then by priority (lower numbers = higher priority)
	if h[i].Time.Equal(h[j].Time) {
		return h[i].Priority < h[j].Priority
	}
	return h[i].Time.Before(h[j].Time)
}

func (h eventHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *eventHeap) Push(x interface{}) {
	*h = append(*h, x.(*Event))
}

func (h *eventHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*h = old[0 : n-1]
	return item
}

// NewEventQueue creates a new event queue
func NewEventQueue() *EventQueue {
	eq := &EventQueue{
		events: make(eventHeap, 0),
	}
	heap.Init(&eq.events)
	return eq
}

// Push adds an event to the queue with O(log n) complexity
func (eq *EventQueue) Push(event *Event) {
	eq.mu.Lock()
	defer eq.mu.Unlock()
	heap.Push(&eq.events, event)
}

// Pop removes and returns the highest priority event from the queue with O(log n) complexity
func (eq *EventQueue) Pop() *Event {
	eq.mu.Lock()
	defer eq.mu.Unlock()
	if eq.events.Len() == 0 {
		return nil
	}
	return heap.Pop(&eq.events).(*Event)
}

// Peek returns the highest priority event without removing it
func (eq *EventQueue) Peek() *Event {
	eq.mu.Lock()
	defer eq.mu.Unlock()
	if eq.events.Len() == 0 {
		return nil
	}
	return eq.events[0]
}

// Size returns the number of events in the queue
func (eq *EventQueue) Size() int {
	eq.mu.Lock()
	defer eq.mu.Unlock()
	return eq.events.Len()
}
