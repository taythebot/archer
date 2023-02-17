package coordinator

import (
	"context"
	"net/http"
)

// Health check
func (c *Client) Health(ctx context.Context) (bool, error) {
	_, err := c.HTTPRequest(ctx, http.MethodGet, "/health", nil)
	if err != nil {
		return false, err
	}

	return true, nil
}
