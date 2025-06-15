package simulator

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"time"

	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
)

// MAX_TICKS sets an upper bound on the number of ticks we'll execute in our main loop
// so we don't simulate unnecessarily long for infinitely-running workflows.
// TODO: make configurable?
const MAX_TICKS = 1_000_000

// FaultConfig configures fault injection behavior
type FaultConfig struct {
	Type        FaultType
	Probability float64
	MinDelay    time.Duration
	MaxDelay    time.Duration
	ErrorMsg    string
}

// SimulatorConfig holds configuration for the simulator
type SimulatorConfig struct {
	Seed               uint64
	TickIncrement      time.Duration
	MaxTicks           int
	FailureProbability float64       // Probability of random failures
	EnableEventJitter  bool          // Add small random delays to events
	FaultConfigs       []FaultConfig // Fault injection configurations
}

// DefaultConfig returns a default configuration
func DefaultConfig() SimulatorConfig {
	return SimulatorConfig{
		Seed:               0, // Will use current time if 0
		TickIncrement:      1 * time.Minute,
		MaxTicks:           MAX_TICKS,
		FailureProbability: 0.0,
		EnableEventJitter:  false,
		FaultConfigs:       []FaultConfig{},
	}
}

// ActivityBehavior defines how an activity should behave during simulation
type ActivityBehavior struct {
	Name         string
	FailureRate  float64
	TimeoutRate  float64
	MinLatency   time.Duration
	MaxLatency   time.Duration
	CustomErrors []error
	ShouldPanic  bool
	ReturnValue  any
	ErrorOnly    bool // If true, activity returns only error; if false, returns (result, error)
}

// ActivityCall represents a call to an activity during simulation
type ActivityCall struct {
	Name      string
	Args      []any
	Timestamp time.Time
	Result    any
	Error     error
	Latency   time.Duration
}

type Simulator struct {
	prng        *rand.Rand
	testenv     *testsuite.TestWorkflowEnvironment
	wfTestSuite testsuite.WorkflowTestSuite

	config        SimulatorConfig
	tickIncrement time.Duration
	invariants    map[string][]Invariant
	eventQueue    *EventQueue
	currentTime   time.Time
	seed          uint64
	wfStarted     chan struct{}
	wfCompleted   chan struct{}

	// Reproducibility tracking
	randomCallCount uint64
	eventHistory    []EventRecord

	// Activity management
	activityBehaviors map[string]ActivityBehavior
	activityCalls     []ActivityCall

	// Fault injection
	faultHistory   []FaultRecord
	activityFaults map[string]error         // Injected activity failures
	networkDelays  map[string]time.Duration // Injected network delays
}

// EventRecord tracks events for reproducibility
type EventRecord struct {
	Tick      int
	Timestamp time.Time
	Type      EventType
	Details   string
}

// FaultRecord tracks injected faults for analysis
type FaultRecord struct {
	Tick      int
	Timestamp time.Time
	Type      FaultType
	Target    string // Activity name, workflow ID, etc.
	Details   string
}

func New(seed uint64) *Simulator {
	return NewWithConfig(SimulatorConfig{
		Seed:          seed,
		TickIncrement: 1 * time.Minute,
		MaxTicks:      MAX_TICKS,
	})
}

// NewWithConfig creates a new simulator with custom configuration
func NewWithConfig(config SimulatorConfig) *Simulator {
	if config.Seed == 0 {
		config.Seed = uint64(time.Now().UnixNano())
	}
	if config.TickIncrement == 0 {
		config.TickIncrement = 1 * time.Minute
	}
	if config.MaxTicks == 0 {
		config.MaxTicks = MAX_TICKS
	}
	sim := &Simulator{
		prng:              rand.New(rand.NewPCG(config.Seed, config.Seed)),
		config:            config,
		invariants:        make(map[string][]Invariant),
		eventQueue:        &EventQueue{events: make([]*Event, 0)},
		currentTime:       time.Now(), // Start time for simulation
		seed:              config.Seed,
		tickIncrement:     config.TickIncrement,
		wfStarted:         make(chan struct{}),
		wfCompleted:       make(chan struct{}),
		randomCallCount:   0,
		eventHistory:      make([]EventRecord, 0),
		faultHistory:      make([]FaultRecord, 0),
		activityBehaviors: make(map[string]ActivityBehavior),
		activityCalls:     make([]ActivityCall, 0),
		activityFaults:    make(map[string]error),
		networkDelays:     make(map[string]time.Duration),
	}
	sim.testenv = sim.wfTestSuite.NewTestWorkflowEnvironment()
	// We want fine grained control of time advancement
	sim.testenv.SetAutoSkipTime(false)
	// main test loop will timeout if it's blocked for 3s by default; that may happen as we're ticking time
	// so we bump it. See: internal_workflow_testsuite.go:944 in temporal go-sdk
	sim.testenv.SetTestTimeout(1 * time.Hour)
	return sim
}

func (s *Simulator) RegisterWorkflow(wf any) {
	s.testenv.RegisterWorkflow(wf)
}

func (s *Simulator) RegisterActivity(activity any) {
	s.testenv.RegisterActivity(activity)
}

// RegisterActivityBehavior defines how an activity should behave during simulation
func (s *Simulator) RegisterActivityBehavior(behavior ActivityBehavior) {
	s.activityBehaviors[behavior.Name] = behavior
}

// RegisterActivityWithBehavior registers an activity with specific failure behavior
func (s *Simulator) RegisterActivityWithBehavior(activity any, behavior ActivityBehavior) {
	// The behavior.Name should match what the workflow calls (e.g., "ValidateCustomer")
	if behavior.Name == "" {
		panic("behavior.Name must be set to match the string used in workflow.ExecuteActivity()")
	}

	// Register behavior for simulation
	s.activityBehaviors[behavior.Name] = behavior

	// Create a wrapper function that will be registered with the expected name
	// Since testsuite doesn't expose RegisterActivityOptions, we use OnActivity mocking
	s.testenv.RegisterActivity(activity)

	// Set up activity mock that respects our behavior
	s.setupActivityMock(behavior.Name, behavior)
}

// setupActivityMock configures the test environment to simulate activity behavior
func (s *Simulator) setupActivityMock(activityName string, behavior ActivityBehavior) {
	// Use the configured return value from behavior
	defaultResult := behavior.ReturnValue

	// Set up basic mocking - activities always receive context as first parameter
	// Use enough mock.Anything parameters to cover all activity signatures
	if behavior.ErrorOnly {
		// Activity returns only error
		s.testenv.OnActivity(activityName, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	} else {
		// Activity returns (result, error)
		s.testenv.OnActivity(activityName, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(defaultResult, nil)
	}
}

// GetActivityCalls returns the history of activity calls
func (s *Simulator) GetActivityCalls() []ActivityCall {
	return s.activityCalls
}

func (s *Simulator) RegisterInvariant(wf any, invariant Invariant) {
	name := name(wf)
	invariants, ok := s.invariants[name]
	if !ok {
		invariants = make([]Invariant, 0)
	}
	invariants = append(invariants, invariant)
	s.invariants[name] = invariants
}

func (s *Simulator) SetTickDuration(d time.Duration) {
	s.tickIncrement = d
	s.config.TickIncrement = d
}

// GetSeed returns the seed used for this simulation
func (s *Simulator) GetSeed() uint64 {
	return s.seed
}

// GetRandomCallCount returns the number of random calls made
func (s *Simulator) GetRandomCallCount() uint64 {
	return s.randomCallCount
}

// GetEventHistory returns the history of events
func (s *Simulator) GetEventHistory() []EventRecord {
	return s.eventHistory
}

// GetFaultHistory returns the history of injected faults
func (s *Simulator) GetFaultHistory() []FaultRecord {
	return s.faultHistory
}

// AddFaultConfig adds a fault injection configuration
func (s *Simulator) AddFaultConfig(config FaultConfig) {
	s.config.FaultConfigs = append(s.config.FaultConfigs, config)
}

// InjectActivityFault manually injects a fault for a specific activity
func (s *Simulator) InjectActivityFault(activityName string, err error) {
	s.activityFaults[activityName] = err
	s.recordFault(FaultActivityError, activityName, fmt.Sprintf("Injected error: %v", err))
}

// InjectNetworkDelay manually injects a network delay
func (s *Simulator) InjectNetworkDelay(target string, delay time.Duration) {
	s.networkDelays[target] = delay
	s.recordFault(FaultNetworkDelay, target, fmt.Sprintf("Injected delay: %v", delay))
}

// recordFault records a fault in the fault history
func (s *Simulator) recordFault(faultType FaultType, target, details string) {
	s.faultHistory = append(s.faultHistory, FaultRecord{
		Tick:      len(s.eventHistory),
		Timestamp: s.currentTime,
		Type:      faultType,
		Target:    target,
		Details:   details,
	})
}

// shouldInjectFault determines if a fault should be injected based on configuration
func (s *Simulator) shouldInjectFault(faultType FaultType) bool {
	for _, config := range s.config.FaultConfigs {
		if config.Type == faultType {
			return s.Random() < config.Probability
		}
	}
	return false
}

// maybeInjectFaults checks and injects random faults during simulation
func (s *Simulator) maybeInjectFaults() {
	// Check for activity failures
	if s.shouldInjectFault(FaultActivityError) {
		activityName := fmt.Sprintf("activity_%d", s.RandomInt(100))
		s.InjectActivityFault(activityName, fmt.Errorf("simulated activity failure"))
	}

	// Check for network delays
	if s.shouldInjectFault(FaultNetworkDelay) {
		target := fmt.Sprintf("network_%d", s.RandomInt(10))
		delay := s.RandomDuration(100*time.Millisecond, 5*time.Second)
		s.InjectNetworkDelay(target, delay)
	}

	// Check for timeouts
	if s.shouldInjectFault(FaultActivityTimeout) {
		activityName := fmt.Sprintf("timeout_activity_%d", s.RandomInt(50))
		s.InjectActivityFault(activityName, fmt.Errorf("activity timeout"))
		s.recordFault(FaultActivityTimeout, activityName, "Simulated timeout")
	}
}

// Random returns a random float64 [0.0, 1.0) - tracked for reproducibility
func (s *Simulator) Random() float64 {
	s.randomCallCount++
	return s.prng.Float64()
}

// RandomInt returns a random integer [0, n) - tracked for reproducibility
func (s *Simulator) RandomInt(n int) int {
	s.randomCallCount++
	return s.prng.IntN(n)
}

// RandomDuration returns a random duration between min and max
func (s *Simulator) RandomDuration(min, max time.Duration) time.Duration {
	s.randomCallCount++
	if min >= max {
		return min
	}
	diff := max - min
	randomNanos := s.prng.Int64N(int64(diff))
	return min + time.Duration(randomNanos)
}

// ShouldInjectFailure returns true if a failure should be injected based on configured probability
func (s *Simulator) ShouldInjectFailure() bool {
	if s.config.FailureProbability <= 0 {
		return false
	}
	return s.Random() < s.config.FailureProbability
}

// validateWorkflowInputs validates that the workflow signature matches provided inputs
func (s *Simulator) validateWorkflowInputs(wf any, input []any) error {
	wfType := reflect.TypeOf(wf)
	
	// Ensure wf is a function
	if wfType.Kind() != reflect.Func {
		return fmt.Errorf("workflow must be a function, got %s", wfType.Kind())
	}
	
	// Calculate expected number of input parameters
	// First parameter should be workflow.Context, remaining are user inputs
	numParams := wfType.NumIn()
	if numParams == 0 {
		return fmt.Errorf("workflow must accept at least workflow.Context as first parameter")
	}
	
	expectedInputs := numParams - 1 // subtract workflow.Context parameter
	actualInputs := len(input)
	
	if actualInputs != expectedInputs {
		return fmt.Errorf("workflow expects %d input parameters, got %d", expectedInputs, actualInputs)
	}
	
	// Type-check input parameters against workflow signature
	for i, inp := range input {
		paramIndex := i + 1 // +1 to skip workflow.Context
		expectedType := wfType.In(paramIndex)
		actualType := reflect.TypeOf(inp)
		
		if !actualType.AssignableTo(expectedType) {
			return fmt.Errorf("parameter %d type mismatch: expected %s, got %s", 
				i, expectedType, actualType)
		}
	}
	
	return nil
}

// scheduleEvent adds an event to the simulation queue
func (s *Simulator) scheduleEvent(event *Event) {
	s.eventQueue.Push(event)
}

// Run is the main simulation loop using deterministic event processing
// Enhanced to accept input parameters for workflows
func (s *Simulator) Run(wf any, input ...any) error {
	// Validate workflow signature and inputs
	if err := s.validateWorkflowInputs(wf, input); err != nil {
		return fmt.Errorf("workflow validation failed: %w", err)
	}
	// Schedule initial workflow start event
	s.scheduleEvent(&Event{
		Time:     s.currentTime,
		Type:     EventWorkflowStart,
		Priority: 0,
		Handler: func() error {
			go func() {
				s.testenv.ExecuteWorkflow(wf, input...)
				close(s.wfCompleted)
			}()
			close(s.wfStarted)
			return nil
		},
	})

	// Schedule regular tick events
	s.scheduleTickEvents()

	// Main event loop
	ticks := 0
	for ticks <= s.config.MaxTicks {
		event := s.eventQueue.Pop()
		if event == nil {
			// No more events, wait for workflow completion or timeout
			select {
			case <-s.wfCompleted:
				fmt.Printf("workflow completed in %d ticks\n", ticks)
				return nil
			default:
				if s.testenv.IsWorkflowCompleted() {
					fmt.Printf("workflow completed in %d ticks\n", ticks)
					return nil
				}
				// Schedule next tick if workflow is still running
				s.scheduleTickEvents()
				continue
			}
		}

		// Process the event
		if err := s.processEvent(event, wf); err != nil {
			return fmt.Errorf("error processing event: %w", err)
		}

		if event.Type == EventTick {
			ticks++
		}

		// Check if workflow completed
		select {
		case <-s.wfCompleted:
			fmt.Printf("workflow completed in %d ticks\n", ticks)
			return nil
		default:
		}

		if s.testenv.IsWorkflowCompleted() {
			fmt.Printf("workflow completed in %d ticks\n", ticks)
			return nil
		}
	}

	fmt.Printf("simulation stopped at max ticks: %d\n", s.config.MaxTicks)
	return nil
}

// scheduleTickEvents schedules the next batch of tick events
func (s *Simulator) scheduleTickEvents() {
	for i := range 10 { // Schedule 10 ticks ahead
		nextTickTime := s.currentTime.Add(time.Duration(i+1) * s.tickIncrement)
		s.scheduleEvent(&Event{
			Time:     nextTickTime,
			Type:     EventTick,
			Priority: 1,
			Handler: func() error {
				s.tick()
				return nil
			},
		})
	}
}

// processEvent handles different types of events
func (s *Simulator) processEvent(event *Event, wf any) error {
	// Record event in history for reproducibility
	s.eventHistory = append(s.eventHistory, EventRecord{
		Tick:      len(s.eventHistory), // Simple tick counter for now
		Timestamp: event.Time,
		Type:      event.Type,
		Details:   fmt.Sprintf("Event processed at %v", event.Time),
	})

	// Update current simulation time
	s.currentTime = event.Time

	// Add jitter if enabled
	if s.config.EnableEventJitter {
		jitter := s.RandomDuration(0, time.Millisecond*100)
		time.Sleep(jitter) // Small delay to simulate real-world timing variations
	}

	// Execute the event handler
	if err := event.Handler(); err != nil {
		return err
	}

	// Perform event-specific actions
	switch event.Type {
	case EventWorkflowStart:
		// Wait for workflow to actually start
		<-s.wfStarted
	case EventTick:
		// Maybe inject faults during tick processing
		s.maybeInjectFaults()
		// Check invariants after each tick
		s.checkInvariants(wf)
	case EventInvariantCheck:
		s.checkInvariants(wf)
	case EventFault:
		// Fault events are handled by their specific handlers
	case EventActivityFailure:
		// Activity failure events are handled by the activity system
	case EventNetworkPartition:
		// Network partition simulation
		s.recordFault(FaultNetworkDelay, "network", "Network partition simulated")
	}

	return nil
}

func (s *Simulator) tick() {
	s.testenv.AdvanceTime(s.tickIncrement)
}

func (s *Simulator) checkInvariants(wf any) {
	name := name(wf)
	invariants := s.invariants[name]

	for _, invariant := range invariants {
		if !invariant.Check() {
			panic(fmt.Sprintf("invariant %s failed!", invariant.Name))
		}
	}
}

func name(obj any) string {
	return reflect.TypeOf(obj).Name()
}
