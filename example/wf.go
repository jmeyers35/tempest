package example

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func Workflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	workflow.Sleep(ctx, 1*time.Hour)
	logger.Info("just slept for an hour")
	return nil
}
