package scheduler

import (
	"context"
	"fmt"

	"github.com/taythebot/archer/pkg/module/nuclei"
)

// Nuclei schedules Nuclei scan tasks
func (s *Scheduler) Nuclei(ctx context.Context, scanId string, targets, templateTypes []string) (tasks []interface{}, err error) {
	// Get number of active workers
	workers := 1
	if activeWorkers, _ := s.Queue.GetWorkers("nuclei"); activeWorkers > 0 {
		workers = activeWorkers
	}

	// Distribute targets
	disTargets := roundRobin(targets, workers)

	// Create tasks
	for _, t := range disTargets {
		// Create task payload
		payload := &nuclei.Payload{
			Scan:          scanId,
			Targets:       t,
			TemplateTypes: templateTypes,
		}

		// Create new asynq task
		asynqTask, err := payload.Create()
		if err != nil {
			return nil, fmt.Errorf("failed to create asynq task: %s", err)
		}

		// Queue task
		task, err := s.queueTask(ctx, "nuclei", scanId, "nuclei", asynqTask)
		if err != nil {
			return nil, err
		}

		// Add to tasks
		tasks = append(tasks, task)
	}

	return tasks, nil
}
