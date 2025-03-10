package simulator

import (
	"fmt"
	"math/rand/v2"
	"time"

	"go.temporal.io/sdk/testsuite"
)

// MAX_TICKS sets an upper bound on the number of ticks we'll execute in our main loop
// so we don't simulate unnecessarily long for infintely-running workflows.
// TODO: make configurable?
const MAX_TICKS = 1_000_000

type Simulator struct {
	prng        *rand.Rand
	testenv     *testsuite.TestWorkflowEnvironment
	wfTestSuite testsuite.WorkflowTestSuite

	tickIncrement time.Duration
}

func New(seed uint64) *Simulator {
	sim := &Simulator{
		// TODO: is this bad?
		prng: rand.New(rand.NewPCG(seed, seed)),
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

func (s *Simulator) SetTickDuration(d time.Duration) {
	s.tickIncrement = d
}

// Run is the main simulation loop. For now, it just ticks time until
// the workflow completes OR MAX_TICKS is reached.
func (s *Simulator) Run(wf any) error {
	// kick off the workflow
	fmt.Println("starting workflow")
	// ExecuteWorkflow is blocking, so unfortunately we have to run in another goroutine. TODO: see if there's a way to kick a workflow in a non-blocking way (may require more patching in the SDK)
	go s.testenv.ExecuteWorkflow(wf)
	fmt.Println("workflow starting, ticking...")
	ticks := 0
	for !s.testenv.IsWorkflowCompleted() && ticks <= MAX_TICKS {
		s.tick()
		ticks += 1
	}
	fmt.Printf("completed in %d ticks\n", ticks)
	return nil
}

func (s *Simulator) tick() {
	fmt.Printf("ticking %v\n", s.tickIncrement)
	s.testenv.AdvanceTime(s.tickIncrement)
}
