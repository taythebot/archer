package types

import "github.com/hibiken/asynq"

// TaskPayload is the abstraction for all task payloads
type TaskPayload interface {
	// Create a new task with payload
	Create() (*asynq.Task, error)

	// ScanId from payload
	ScanId() string
}

// TODO: rename struct or move to Elasticsearch module

// TaskResult is the data into Elasticsearch indexer
type TaskResult struct {
	ID  string      // ID is the document ID
	Doc interface{} // Doc is the document body
}
