package nuclei

import (
	"github.com/hibiken/asynq"
)

// Create creates an Asynq task from the Payload
func (p *Payload) Create() (*asynq.Task, error) {
	// Marshal to JSON
	payload, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	// Create new task
	return asynq.NewTask("nuclei", payload), nil
}

// ScanId from payload
func (p *Payload) ScanId() string {
	return p.Scan
}
