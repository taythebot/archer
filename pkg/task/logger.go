package task

import (
	"context"
	"errors"

	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
)

// createLogger creates a child logger with the task metadata
func createLogger(ctx context.Context, scanId, taskId, module string) (*logrus.Entry, error) {
	// Get worker ID
	workerId := ctx.Value("worker")
	if workerId == "" {
		return nil, errors.New("failed to get worker ID")
	}

	// Get queue name
	queue, ok := asynq.GetQueueName(ctx)
	if !ok {
		return nil, errors.New("failed to get queue name")
	}

	// Get retry count
	retryCount, ok := asynq.GetRetryCount(ctx)
	if !ok {
		return nil, errors.New("failed to get retry count")
	}

	// Get max retry count
	maxRetry, ok := asynq.GetMaxRetry(ctx)
	if !ok {
		return nil, errors.New("failed to get max retry count")
	}

	// Create child logger
	return logrus.WithFields(
		logrus.Fields{
			"worker":     workerId,
			"module":     module,
			"scan":       scanId,
			"task":       taskId,
			"queue":      queue,
			"retryCount": retryCount,
			"maxRetry":   maxRetry,
		},
	), nil
}
