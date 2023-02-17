package coordinator

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"
)

// Client for Coordinator
type Client struct {
	URL        string
	WorkerID   string
	httpClient *http.Client
	Debug      bool
}

// New creates a new Coordinator client
func New(url, workerID string, debug bool) *Client {
	return &Client{
		URL:      url,
		WorkerID: workerID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		Debug: debug,
	}
}

// HTTPRequest creates an HTTP request and executes it
func (c *Client) HTTPRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	// Create request
	req, err := http.NewRequestWithContext(ctx, method, c.URL+path, body)
	if err != nil {
		return nil, err
	}

	// Add content-type header
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	// Add custom headers
	if c.WorkerID != "" {
		req.Header.Add("User-Agent", "archer-worker/1.0.0")
		req.Header.Add("X-Worker-ID", c.WorkerID)
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
		if c.Debug {
			defer resp.Body.Close()
			if respBody, err := ioutil.ReadAll(resp.Body); err == nil {
				log.Debugf("Coordinator Response: %s", respBody)
			}
		}

		requestURL, _ := url.QueryUnescape(req.URL.String())
		return resp, fmt.Errorf("unexpected status code %d received from %s", resp.StatusCode, requestURL)
	}

	return resp, nil
}
