package main

import (
	"fmt"
	"time"

	"github.com/jmeyers35/tempest/example"
	"github.com/jmeyers35/tempest/pkg/simulator"
)

func main() {
	sim := simulator.New(12141997)
	sim.SetTickDuration(1 * time.Minute)

	wf := &example.Workflow{}

	sim.RegisterWorkflow(wf.DoALongBankTransaction)
	sim.RegisterInvariant(wf.DoALongBankTransaction, simulator.Invariant{
		Name: "balance must always be positive",
		Check: func() bool {
			return wf.Balance >= 0
		},
	})
	fmt.Println("starting simulator")
	sim.Run(wf.DoALongBankTransaction)
}
