package coordinator

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client for Coordinator
type Client struct {
	Url        string
	WorkerId   string
	httpClient *http.Client
}

// New creates a new Coordinator client
func New(url, workerId string) *Client {
	return &Client{
		Url:      url,
		WorkerId: workerId,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// HTTPRequest creates an HTTP request and executes it
func (c *Client) HTTPRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	// Create request
	req, err := http.NewRequestWithContext(ctx, method, c.Url+path, body)
	if err != nil {
		return nil, err
	}

	// Add content-type header
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	// Add custom headers
	if c.WorkerId != "" {
		req.Header.Add("User-Agent", "archer-worker/1.0.0")
		req.Header.Add("X-Worker-ID", c.WorkerId)
	} else {
		req.Header.Add("User-Agent", "archer-cli/1.0.0")
	}

	// Perform request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		requestURL, _ := url.QueryUnescape(req.URL.String())
		return resp, fmt.Errorf("unexpected status code %d received from %s", resp.StatusCode, requestURL)
	}

	return resp, nil
}
