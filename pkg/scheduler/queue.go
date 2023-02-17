package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/taythebot/archer/pkg/model"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgtype"
)

// queueTask creates a new task in the database and then queues it
func (s *Scheduler) queueTask(ctx context.Context, queue, scanId, module string, asynqTask *asynq.Task) (*model.Tasks, error) {
	// Convert payload to JSONB
	var payloadJsonB pgtype.JSONB
	if err := payloadJsonB.Set(asynqTask.Payload()); err != nil {
		return nil, fmt.Errorf("failed to convert task payload to JSONB: %s", err)
	}

	// Create task in database
	task := &model.Tasks{
		ScanID:  scanId,
		Module:  module,
		Payload: &payloadJsonB,
	}
	if err := s.DB.Create(task).Error; err != nil {
		return nil, fmt.Errorf("failed to create task in database: %s", err)
	}

	// Queue task
	if _, err := s.Queue.Enqueue(ctx, queue, asynqTask, task.ID, 24*time.Hour); err != nil {
		return nil, fmt.Errorf("failed to queue task: %s", err)
	}

	return task, nil
}
