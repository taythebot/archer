package task

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hibiken/asynq"
)

// ProcessScanTask handles scan tasks from Asynq
func (th *TaskHandler) ProcessScanTask(ctx context.Context, task *asynq.Task) error {
	// Get handler
	handler, pattern, err := th.Handler(task)
	if err != nil {
		return err
	}

	// Get payload struct
	payload := handler.Payload()

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

	// Start Elasticsearch indexer
	logger.Info("Creating Elasticsearch indexer")
	indexer, err := th.Elasticsearch.NewIndexer(logger)
	if err != nil {
		return fmt.Errorf("failed to create Elasticsearch indexer: %s", err)
	}

	// Run task
	total, err := handler.ProcessTask(ctx, taskId, payload, indexer, logger)
	if err != nil {
		return err
	}

	// Close indexer
	logger.Debug("Closing Elasticsearch indexer")
	if err := indexer.Close(ctx); err != nil {
		logger.Errorf("Failed to close Elasticsearch indexer: %s", err)
	}

	// Check total indexed and returned
	if total != indexer.TotalIndexed {
		return fmt.Errorf("total indexed and returned is not the same")
	}

	logger.Infof("Task finished with %d results", total)

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
