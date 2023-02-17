package elasticsearch

import jsoniter "github.com/json-iterator/go"

// OpenPit is the response body for an open PIT request
type OpenPit struct {
	Id string `json:"id"`
}

// SearchBody is the request body for a search request
type SearchBody struct {
	Size        int           `json:"size,omitempty"`
	Query       interface{}   `json:"query,omitempty"`
	Pit         SearchBodyPit `json:"pit,omitempty"`
	Sort        []interface{} `json:"sort,omitempty"`
	SearchAfter []interface{} `json:"search_after,omitempty"`
}

// SearchBodyPit is the pit part of a search request
type SearchBodyPit struct {
	ID        string `json:"id"`
	KeepAlive string `json:"keep_alive"`
}

// SearchResult is the response body for a search request
type SearchResult struct {
	PitId    string `json:"pit_id,omitempty"`
	Took     int    `json:"took"`
	TimedOut bool   `json:"timed_out"`
	Hits     struct {
		Total struct {
			Value int `json:"total"`
		} `json:"total"`
		Hits []struct {
			Index  string              `json:"_index"`
			ID     string              `json:"_id"`
			Source jsoniter.RawMessage `json:"_source"`
			Sort   []interface{}       `json:"sort,omitempty"`
		} `json:"hits"`
	} `json:"hits"`
}

// ApiError is the error response body from Elasticsearch
type ApiError struct {
	Error struct {
		RootCause []struct {
			Type   string `yaml:"type"`
			Reason string `yaml:"reason"`
		} `yaml:"root_cause"`
		Type   string `yaml:"type"`
		Reason string `yaml:"reason"`
	} `yaml:"error"`
	Status int `json:"status"`
}
