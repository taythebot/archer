package coordinator

import (
	"bytes"
	"context"
	"net/http"

	"github.com/taythebot/archer/cmd/coordinator/form"
)

// TaskComplete ...
type TaskComplete struct {
	Results int `json:"results"`
}

// TaskStarted marks a task as "active"
func (c *Client) TaskStarted(ctx context.Context, taskId string) error {
	_, err := c.HTTPRequest(ctx, http.MethodPost, "/tasks/"+taskId+"/started", bytes.NewBufferString("{}"))
	if err != nil {
		return err
	}

	return nil
}

// TaskCompleted marks a task as "completed"
func (c *Client) TaskCompleted(ctx context.Context, taskId string, results int) error {
	// Create request body
	body, err := json.Marshal(form.CompletedTask{Results: results})
	if err != nil {
		return err
	}

	_, err = c.HTTPRequest(ctx, http.MethodPost, "/tasks/"+taskId+"/completed", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	return nil
}

// TaskFailed marks a task as "failed"
func (c *Client) TaskFailed(ctx context.Context, taskId string) error {
	_, err := c.HTTPRequest(ctx, http.MethodPost, "/tasks/"+taskId+"/failed", bytes.NewBufferString("{}"))
	if err != nil {
		return err
	}

	return nil
}
