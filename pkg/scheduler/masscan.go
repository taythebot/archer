package scheduler

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/taythebot/archer/pkg/model"
	"github.com/taythebot/archer/pkg/module/masscan"
)

// Masscan schedules Masscan scan tasks
func (s *Scheduler) Masscan(ctx context.Context, scanId string, targets []string, ports []uint16) (tasks []*model.Tasks, err error) {
	// Get number of active workers
	workers := 1
	if activeWorkers, _ := s.Queue.GetWorkers("masscan"); activeWorkers > 0 {
		workers = activeWorkers
	}

	// Distribute targets
	disTargets := roundRobin(targets, workers)

	// Create seed
	seed := time.Now().UnixNano()

	// Get total tasks count
	totalTasks := strconv.Itoa(len(disTargets))

	// Create tasks
	for index, t := range disTargets {
		// Create task payload
		payload := &masscan.Payload{
			Scan:    scanId,
			Targets: t,
			Ports:   ports,
			Shard:   strconv.Itoa(index+1) + "/" + totalTasks,
			Seed:    seed,
		}

		// Create new asynq task
		asynqTask, err := payload.Create()
		if err != nil {
			return nil, fmt.Errorf("failed to create asynq task: %s", err)
		}

		// Queue task
		task, err := s.queueTask(ctx, "masscan", scanId, "masscan", asynqTask)
		if err != nil {
			return nil, err
		}

		// Add to tasks
		tasks = append(tasks, task)
	}

	return tasks, nil
}
