package task

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/taythebot/archer/pkg/types"

	"github.com/hibiken/asynq"
)

// ProcessSchedulerTask handles scheduler tasks from Asynq
func (th *TaskHandler) ProcessSchedulerTask(ctx context.Context, task *asynq.Task) error {
	// Check for Scheduler
	if th.Scheduler == nil {
		return errors.New("scheduler not found in Task client")
	}

	// Get handler
	handler, pattern, err := th.Handler(task)
	if err != nil {
		return err
	}

	// Get payload struct
	payload := types.SchedulerPayload{}

	// Unmarshal payload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %s", err)
	}

	// Get module name
	module := handler.Name()
	ctx = context.WithValue(ctx, "module", module)

	// Get scan ID
	scanId := payload.ScanId()
	ctx = context.WithValue(ctx, "scan", scanId)

	// Get task ID
	taskId, ok := asynq.GetTaskID(ctx)
	if !ok {
		return errors.New("failed to get task id")
	}

	// Create logger
	logger, err := createLogger(ctx, scanId, taskId, module)
	if err != nil {
		return fmt.Errorf("failed to create logger: %s", err)
	}

	// Add logger to context
	ctx = context.WithValue(ctx, "logger", logger)

	logger.Debugf("Matched task handler pattern '%s'", pattern)

	// Signal Coordinator for task start
	if err := th.Coordinator.TaskStarted(ctx, taskId); err != nil {
		return fmt.Errorf("failed to notify Coordinator of task start: %s", err)
	}

	// Run task
	total, err := handler.Schedule(ctx, payload, th.Scheduler, th.Elasticsearch, logger)
	if err != nil {
		return err
	}

	logger.Infof("Successfully scheduled %d new tasks", total)

	// Save total results to Queue
	totalStr := strconv.Itoa(int(total))
	if _, err = task.ResultWriter().Write([]byte(totalStr)); err != nil {
		logger.Errorf("Failed to write to Task results: %s", err)
	}

	// Signal Coordinator for task finished
	if err := th.Coordinator.TaskCompleted(ctx, taskId, int(total)); err != nil {
		return fmt.Errorf("failed to notify Coordinator of task completion: %s", err)
	}

	return nil
}
