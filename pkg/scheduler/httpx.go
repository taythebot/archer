package scheduler

import (
	"context"
	"fmt"

	"github.com/taythebot/archer/pkg/module/httpx"
)

// Httpx schedules Httpx scan tasks
func (s *Scheduler) Httpx(ctx context.Context, scanId string, targets []string) (tasks []interface{}, err error) {
	// Get number of active workers
	workers := 1
	if activeWorkers, _ := s.Queue.GetWorkers("httpx"); activeWorkers > 0 {
		workers = activeWorkers
	}

	// Distribute targets
	disTargets := roundRobin(targets, workers)

	// Create tasks
	for _, t := range disTargets {
		// Create task payload
		payload := &httpx.Payload{
			Scan:    scanId,
			Targets: t,
		}

		// Create new asynq task
		asynqTask, err := payload.Create()
		if err != nil {
			return nil, fmt.Errorf("failed to create asynq task: %s", err)
		}

		// Queue task
		task, err := s.queueTask(ctx, "httpx", scanId, "httpx", asynqTask)
		if err != nil {
			return nil, err
		}

		// Add to tasks
		tasks = append(tasks, task)
	}

	return tasks, nil
}
