package types

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
)

// Scheduler interface
type Scheduler interface {
	Httpx(context.Context, string, []string) ([]interface{}, error)
	Nuclei(context.Context, string, []string, []string) ([]interface{}, error)
}

// SchedulerPayload is the payload for tasks scheduling future stages
type SchedulerPayload struct {
	Scan           string `json:"scan"`            // Scan ID
	Task           string `json:"task"`            // Task ID
	PreviousModule string `json:"previous_module"` // PreviousModule
	Module         string `json:"module"`          // Module
}

// Create creates an Asynq task from the Payload
func (p *SchedulerPayload) Create() (*asynq.Task, error) {
	// Marshal to JSON
	payload, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	// Create new task
	return asynq.NewTask(p.Module, payload), nil
}

// ScanId from payload
func (p *SchedulerPayload) ScanId() string {
	return p.Scan
}
