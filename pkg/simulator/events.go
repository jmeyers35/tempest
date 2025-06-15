package simulator

import (
	"time"
)

// Event represents a scheduled event in the simulation
type Event struct {
	Time     time.Time
	Type     EventType
	Handler  func() error
	Priority int // Lower numbers have higher priority
}

// EventType categorizes different types of events
type EventType int

const (
	EventTick EventType = iota
	EventFault
	EventWorkflowStart
	EventWorkflowComplete
	EventInvariantCheck
	EventActivityFailure
	EventTimeoutInjection
	EventNetworkPartition
)

// String returns the string representation of EventType
func (et EventType) String() string {
	switch et {
	case EventTick:
		return "Tick"
	case EventFault:
		return "Fault"
	case EventWorkflowStart:
		return "WorkflowStart"
	case EventWorkflowComplete:
		return "WorkflowComplete"
	case EventInvariantCheck:
		return "InvariantCheck"
	case EventActivityFailure:
		return "ActivityFailure"
	case EventTimeoutInjection:
		return "TimeoutInjection"
	case EventNetworkPartition:
		return "NetworkPartition"
	default:
		return "Unknown"
	}
}

// FaultType represents different types of faults that can be injected
type FaultType int

const (
	FaultActivityTimeout FaultType = iota
	FaultActivityError
	FaultNetworkDelay
	FaultWorkerCrash
	FaultDatabaseUnavailable
)

// String returns the string representation of FaultType
func (ft FaultType) String() string {
	switch ft {
	case FaultActivityTimeout:
		return "ActivityTimeout"
	case FaultActivityError:
		return "ActivityError"
	case FaultNetworkDelay:
		return "NetworkDelay"
	case FaultWorkerCrash:
		return "WorkerCrash"
	case FaultDatabaseUnavailable:
		return "DatabaseUnavailable"
	default:
		return "Unknown"
	}
}
