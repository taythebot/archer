package elasticsearch

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// OpenPit creates a new PIT for an index
func (c *Client) OpenPit(ctx context.Context, keepAlive string) (string, error) {
	// Create new open PIT request
	pit := esapi.OpenPointInTimeRequest{
		Index:     []string{c.Index},
		KeepAlive: keepAlive,
	}

	// Execute request
	resp, err := pit.Do(ctx, c.Elasticsearch)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check for response error
	if resp.IsError() {
		// Decode API error
		var e ApiError
		if err := json.NewDecoder(resp.Body).Decode(e); err != nil {
			return "", fmt.Errorf("failed to unmarshal error response body: %s", err)
		}

		return "", fmt.Errorf("failed to perform open PIT request: %s", e.Error)
	}

	// Decode response body
	var result OpenPit
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to unmarshal resposne body: %s", err)
	}

	return result.Id, nil
}
