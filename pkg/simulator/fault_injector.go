package simulator

import (
	"fmt"
	"time"
)

// FaultInjector manages fault injection configurations and execution
type FaultInjector struct {
	configs         []FaultConfig
	activityFaults  map[string]error
	networkDelays   map[string]time.Duration
	faultHistory    []FaultRecord
	randomGenerator RandomGenerator
}

// RandomGenerator provides random number generation for fault injection
type RandomGenerator interface {
	Float64() float64
	IntN(n int) int
	Int64N(n int64) int64
}

// NewFaultInjector creates a new fault injector
func NewFaultInjector(rng RandomGenerator) *FaultInjector {
	return &FaultInjector{
		configs:         make([]FaultConfig, 0),
		activityFaults:  make(map[string]error),
		networkDelays:   make(map[string]time.Duration),
		faultHistory:    make([]FaultRecord, 0),
		randomGenerator: rng,
	}
}

// AddFaultConfig adds a fault injection configuration
func (fi *FaultInjector) AddFaultConfig(config FaultConfig) error {
	if config.Probability < 0 || config.Probability > 1 {
		return fmt.Errorf("fault probability must be between 0 and 1, got %f", config.Probability)
	}

	if config.MinDelay < 0 || config.MaxDelay < 0 {
		return fmt.Errorf("fault delays must be non-negative")
	}

	if config.MinDelay > config.MaxDelay {
		return fmt.Errorf("minimum delay (%v) cannot be greater than maximum delay (%v)",
			config.MinDelay, config.MaxDelay)
	}

	fi.configs = append(fi.configs, config)
	return nil
}

// ShouldInjectFault determines if a fault should be injected based on configuration
func (fi *FaultInjector) ShouldInjectFault(faultType FaultType) bool {
	for _, config := range fi.configs {
		if config.Type == faultType {
			return fi.randomGenerator.Float64() < config.Probability
		}
	}
	return false
}

// InjectActivityFault manually injects a fault for a specific activity
func (fi *FaultInjector) InjectActivityFault(activityName string, err error) {
	fi.activityFaults[activityName] = err
	fi.recordFault(FaultActivityError, activityName, fmt.Sprintf("Injected error: %v", err))
}

// InjectNetworkDelay manually injects a network delay
func (fi *FaultInjector) InjectNetworkDelay(target string, delay time.Duration) error {
	if delay < 0 {
		return fmt.Errorf("network delay must be non-negative, got %v", delay)
	}

	fi.networkDelays[target] = delay
	fi.recordFault(FaultNetworkDelay, target, fmt.Sprintf("Injected delay: %v", delay))
	return nil
}

// ProcessFaultInjection checks and injects faults based on configuration
func (fi *FaultInjector) ProcessFaultInjection() {
	// Process activity failures
	if fi.ShouldInjectFault(FaultActivityError) {
		activityName := fi.generateActivityTarget()
		fi.InjectActivityFault(activityName, fmt.Errorf("simulated activity failure"))
	}

	// Process network delays
	if fi.ShouldInjectFault(FaultNetworkDelay) {
		target := fi.generateNetworkTarget()
		delay := fi.generateRandomDelay(100*time.Millisecond, 5*time.Second)
		fi.InjectNetworkDelay(target, delay)
	}

	// Process activity timeouts
	if fi.ShouldInjectFault(FaultActivityTimeout) {
		activityName := fi.generateTimeoutTarget()
		fi.InjectActivityFault(activityName, fmt.Errorf("activity timeout"))
		fi.recordFault(FaultActivityTimeout, activityName, "Simulated timeout")
	}
}

// GetActivityFault returns any injected fault for the given activity
func (fi *FaultInjector) GetActivityFault(activityName string) error {
	return fi.activityFaults[activityName]
}

// GetNetworkDelay returns any injected delay for the given target
func (fi *FaultInjector) GetNetworkDelay(target string) time.Duration {
	return fi.networkDelays[target]
}

// GetFaultHistory returns the history of injected faults
func (fi *FaultInjector) GetFaultHistory() []FaultRecord {
	return fi.faultHistory
}

// ClearFaults removes all injected faults
func (fi *FaultInjector) ClearFaults() {
	fi.activityFaults = make(map[string]error)
	fi.networkDelays = make(map[string]time.Duration)
}

// recordFault records a fault in the fault history
func (fi *FaultInjector) recordFault(faultType FaultType, target, details string) {
	fi.faultHistory = append(fi.faultHistory, FaultRecord{
		Tick:      len(fi.faultHistory), // Simple incrementing tick for now
		Timestamp: time.Now(),
		Type:      faultType,
		Target:    target,
		Details:   details,
	})
}

// generateActivityTarget creates a deterministic activity target name
func (fi *FaultInjector) generateActivityTarget() string {
	const maxActivities = 100
	return fmt.Sprintf("activity_%d", fi.randomGenerator.IntN(maxActivities))
}

// generateNetworkTarget creates a deterministic network target name
func (fi *FaultInjector) generateNetworkTarget() string {
	const maxNetworkTargets = 10
	return fmt.Sprintf("network_%d", fi.randomGenerator.IntN(maxNetworkTargets))
}

// generateTimeoutTarget creates a deterministic timeout target name
func (fi *FaultInjector) generateTimeoutTarget() string {
	const maxTimeoutTargets = 50
	return fmt.Sprintf("timeout_activity_%d", fi.randomGenerator.IntN(maxTimeoutTargets))
}

// generateRandomDelay generates a random delay between min and max
func (fi *FaultInjector) generateRandomDelay(min, max time.Duration) time.Duration {
	if min >= max {
		return min
	}
	diff := max - min
	randomNanos := fi.randomGenerator.Int64N(int64(diff))
	return min + time.Duration(randomNanos)
}
