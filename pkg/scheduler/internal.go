package scheduler

import (
	"context"
	"fmt"

	"github.com/taythebot/archer/pkg/model"
	"github.com/taythebot/archer/pkg/types"
)

// Internal schedules an internal Scheduler task
func (s *Scheduler) Internal(ctx context.Context, module, scanId, taskId, previousModule string) (*model.Tasks, error) {
	// Create new task payload
	payload := types.SchedulerPayload{
		Scan:           scanId,
		Task:           taskId,
		PreviousModule: previousModule,
		Module:         module,
	}

	// Create new asynq task
	asynqTask, err := payload.Create()
	if err != nil {
		return nil, fmt.Errorf("failed to create asynq task: %s", err)
	}

	// Queue task
	return s.queueTask(ctx, "scheduler", scanId, "scheduler", asynqTask)
}
