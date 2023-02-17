package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/taythebot/archer/pkg/coordinator"
	"github.com/taythebot/archer/pkg/types"

	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
)

// Server for queue
type Server struct {
	Asynq *asynq.Server
}

// NewServer creates a new queue server
func NewServer(config types.RedisServerConfig, queues []string, concurrency int, workerId string, coordinator *coordinator.Client, debug bool) *Server {
	// Create base config
	asynqConfig := asynq.Config{
		Concurrency: concurrency,
		Logger:      logrus.New(),
		BaseContext: func() context.Context {
			return context.WithValue(context.Background(), "worker", workerId)
		},
		ErrorHandler: errorHandler(coordinator, workerId),
		HealthCheckFunc: func(err error) {
			if err != nil {
				logrus.WithField("worker", workerId).Errorf("Failed to send heartbeat: %s", err)
			}
		},
		Queues: make(map[string]int, len(queues)),
	}

	// Enable debug logs
	if debug {
		asynqConfig.LogLevel = asynq.DebugLevel
	}

	// Add queues
	for _, queue := range queues {
		asynqConfig.Queues[queue] = 1
	}

	// Set heartbeat interval
	if config.Heartbeat > 0 {
		asynqConfig.HealthCheckInterval = time.Duration(config.Heartbeat) * time.Second
	}

	return &Server{
		Asynq: asynq.NewServer(
			asynq.RedisClientOpt{
				Addr:     config.Host,
				Username: config.Username,
				Password: config.Password,
				DB:       config.Database,
			},
			asynqConfig,
		),
	}
}

// Run the handler function and process tasks
func (s *Server) Run(handlerFunc asynq.HandlerFunc) error {
	return s.Asynq.Run(handlerFunc)
}

// errorHandler handles errors from the queue
func errorHandler(coordinator *coordinator.Client, workerId string) asynq.ErrorHandlerFunc {
	return func(ctx context.Context, task *asynq.Task, err error) {
		// Get task ID
		taskId, ok := asynq.GetTaskID(ctx)
		if !ok {
			logrus.Errorf("Failed to get task ID in queue errorHandler")
			return
		}

		// Get retries
		retryCount, _ := asynq.GetRetryCount(ctx)
		maxRetry, _ := asynq.GetMaxRetry(ctx)

		// Create logger
		logger := logrus.WithFields(
			logrus.Fields{
				"worker":     workerId,
				"task":       taskId,
				"retryCount": retryCount,
				"maxRetry":   maxRetry,
			},
		)

		// Check retries
		if retryCount >= maxRetry {
			err = fmt.Errorf("retry exhausted for task %s: %w", task.Type, err)
		}

		// Log error
		logger.Error(err)

		// Notify Coordinator
		if err := coordinator.TaskFailed(ctx, taskId); err != nil {
			err = fmt.Errorf("failed to notify Coordinator of task failure: %s", err)

			logger.Error(err.Error())
		}
	}
}
