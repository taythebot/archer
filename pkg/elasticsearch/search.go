package elasticsearch

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// SearchAfter allows paginating results using the PIT and search_after parameters
func (c *Client) SearchAfter(ctx context.Context, query string, size int, source []string, searchAfter []interface{}, sort, pitId, pitKeepAlive string) (*SearchResult, error) {
	// Validate parameters
	if pitId == "" {
		return nil, errors.New("pitId is missing")
	} else if pitKeepAlive == "" {
		return nil, errors.New("pitKeepAlive is missing")
	} else if sort == "" {
		return nil, errors.New("sort is required")
	}

	// Create base options
	opts := []func(*esapi.SearchRequest){
		c.Elasticsearch.Search.WithContext(ctx),
		c.Elasticsearch.Search.WithQuery(query),
		c.Elasticsearch.Search.WithSort(sort),
	}

	// Add size
	if size > 0 {
		opts = append(opts, c.Elasticsearch.Search.WithSize(size))
	}

	// Add source
	if len(source) > 0 {
		opts = append(opts, c.Elasticsearch.Search.WithSource(source...))
	}

	// Create search body
	body := SearchBody{
		Pit: SearchBodyPit{
			ID:        pitId,
			KeepAlive: pitKeepAlive,
		},
		SearchAfter: searchAfter,
	}

	// Marshal to JSON
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %s", err)
	}

	// Add to options
	opts = append(opts, c.Elasticsearch.Search.WithBody(bytes.NewReader(bodyBytes)))

	// Perform search request
	resp, err := c.Elasticsearch.Search(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to perform search request: %s", err)
	}
	defer resp.Body.Close()

	// Check for response error
	if resp.IsError() {
		// Decode API error
		var e ApiError
		if err := json.NewDecoder(resp.Body).Decode(e); err != nil {
			return nil, fmt.Errorf("failed to unmarshal error response body: %s", err)
		}

		return nil, fmt.Errorf("failed to perform search request: %s", e.Error)
	}

	// Decode response body
	result := &SearchResult{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal resposne body: %s", err)
	}

	return result, nil
}

// Search index with Lucene query string syntax
func (c *Client) Search(ctx context.Context, query string, size int, source []string) (*SearchResult, error) {
	// Create base options
	opts := []func(*esapi.SearchRequest){
		c.Elasticsearch.Search.WithIndex(c.Index),
		c.Elasticsearch.Search.WithContext(ctx),
		c.Elasticsearch.Search.WithQuery(query),
		c.Elasticsearch.Search.WithTrackTotalHits(true),
	}

	// Add size
	if size > 0 {
		opts = append(opts, c.Elasticsearch.Search.WithSize(size))
	}

	// Add source
	if len(source) > 0 {
		opts = append(opts, c.Elasticsearch.Search.WithSource(source...))
	}

	// Perform search request
	resp, err := c.Elasticsearch.Search(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to perform search request: %s", err)
	}
	defer resp.Body.Close()

	// Check for response error
	if resp.IsError() {
		// Decode API error
		var e ApiError
		if err := json.NewDecoder(resp.Body).Decode(e); err != nil {
			return nil, fmt.Errorf("failed to unmarshal error response body: %s", err)
		}

		return nil, fmt.Errorf("failed to perform search request: %s", e.Error)
	}

	// Decode response body
	result := &SearchResult{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal resposne body: %s", err)
	}

	return result, nil
}
