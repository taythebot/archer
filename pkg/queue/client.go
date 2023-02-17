package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/taythebot/archer/pkg/types"

	"github.com/hibiken/asynq"
)

// Client for queue
type Client struct {
	Client        *asynq.Client
	Inspector     *asynq.Inspector
	TaskRetention time.Duration
}

// NewClient creates a new queue client
func NewClient(config types.RedisClientConfig) (*Client, error) {
	client := &Client{}

	// Parse task duration
	if config.TaskRetention != "" {
		duration, err := time.ParseDuration(config.TaskRetention)
		if err != nil {
			return nil, fmt.Errorf("invalid task duration: %s", err)
		}

		client.TaskRetention = duration
	}

	// Create redis config
	redisOpt := asynq.RedisClientOpt{
		Addr:     config.Host,
		Username: config.Username,
		Password: config.Password,
		DB:       config.Database,
	}

	// Create client and inspector
	client.Client = asynq.NewClient(redisOpt)
	client.Inspector = asynq.NewInspector(redisOpt)

	return client, nil
}

// Enqueue will queue a new task
func (c *Client) Enqueue(ctx context.Context, queue string, task *asynq.Task, id string, timeout time.Duration) (*asynq.TaskInfo, error) {
	// Create base options
	opts := []asynq.Option{
		asynq.Queue(queue),
		asynq.Timeout(timeout),
	}

	// Add task duration
	if c.TaskRetention != 0 {
		opts = append(opts, asynq.Retention(c.TaskRetention))
	}

	// Add task id
	if id != "" {
		opts = append(opts, asynq.TaskID(id))
	}

	return c.Client.EnqueueContext(ctx, task, opts...)
}

// GetWorkers gets active workers for a queue
func (c *Client) GetWorkers(queue string) (int, error) {
	// Get servers
	servers, err := c.Inspector.Servers()
	if err != nil {
		return 0, err
	}

	// Parse servers
	var workers int
	for _, server := range servers {
		// Check for queue
		for q := range server.Queues {
			if q == queue {
				workers += server.Concurrency
				break
			}
		}
	}

	return workers, nil
}
