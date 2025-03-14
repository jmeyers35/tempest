package example

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

type Workflow struct {
	Balance float64
}

func (wf *Workflow) DoALongBankTransaction(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	workflow.Sleep(ctx, 1*time.Hour)
	logger.Info("just slept for an hour")
	wf.Balance -= 100
	return nil
}
