package elasticsearch

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/taythebot/archer/pkg/types"

	"github.com/elastic/go-elasticsearch/v8"
)

// Client for Elasticsearch
type Client struct {
	Elasticsearch *elasticsearch.Client
	Index         string
	FlushBytes    int
	FlushInterval int
	Pipeline      string
}

// New creates a new Elasticsearch client
func New(config types.ElasticConfig, debug bool) (*Client, error) {
	// Create new client
	client, err := elasticsearch.NewClient(
		elasticsearch.Config{
			Addresses:         config.Hosts,
			Username:          config.Username,
			Password:          config.Password,
			EnableDebugLogger: debug,
			RetryOnStatus:     []int{502, 503, 504, 429},
			MaxRetries:        5,
			Transport: &http.Transport{
				MaxIdleConnsPerHost:   10,
				ResponseHeaderTimeout: time.Second,
				DialContext:           (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
				TLSClientConfig: &tls.Config{
					MinVersion:         tls.VersionTLS12,
					InsecureSkipVerify: true,
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	// Check if index exists
	resp, err := client.Indices.Exists([]string{config.Index})
	if err != nil {
		return nil, err
	} else if resp.IsError() {
		return nil, fmt.Errorf("index '%s' not found", config.Index)
	}
	if err := resp.Body.Close(); err != nil {
		return nil, err
	}

	return &Client{
		Elasticsearch: client,
		Index:         config.Index,
		FlushBytes:    config.Bulk.FlushBytes,
		FlushInterval: config.Bulk.FlushInterval,
		Pipeline:      config.Bulk.Pipeline,
	}, nil
}
