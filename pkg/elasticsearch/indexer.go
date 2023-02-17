package elasticsearch

import (
	"bytes"
	"context"
	"sync/atomic"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esutil"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

// Indexer for Elasticsearch
type Indexer struct {
	Client       esutil.BulkIndexer
	Logger       *logrus.Entry
	TotalIndexed uint32
}

// ScriptedUpdate is the request body for a scripted updated
type ScriptedUpdate struct {
	ScriptedUpsert bool     `json:"scripted_upsert"`
	Script         Script   `json:"script"`
	Upsert         struct{} `json:"upsert"`
}

// Script for scripted updates
type Script struct {
	Source string      `json:"source"`
	Lang   string      `json:"lang"`
	Params interface{} `json:"params"`
}

var json = jsoniter.ConfigFastest

// NewIndexer creates a new indexer
func (c *Client) NewIndexer(logger *logrus.Entry) (*Indexer, error) {
	// Create new indexer
	indexer, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         c.Index,
		Client:        c.Elasticsearch,
		FlushBytes:    c.FlushBytes,
		FlushInterval: time.Duration(c.FlushInterval) * time.Second,
		Pipeline:      c.Pipeline,
	})
	if err != nil {
		return nil, err
	}

	return &Indexer{Client: indexer, Logger: logger}, nil
}

// Add adds a document to the processor
func (i *Indexer) Add(ctx context.Context, id string, document interface{}, script string) error {
	// Create new item
	item := esutil.BulkIndexerItem{
		Action:     "index",
		DocumentID: id,
		OnSuccess: func(_ context.Context, _ esutil.BulkIndexerItem, _ esutil.BulkIndexerResponseItem) {
			atomic.AddUint32(&i.TotalIndexed, 1)
		},
		OnFailure: func(_ context.Context, _ esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
			if err != nil {
				i.Logger.Errorf("Failed to index document: %s", err)
			} else {
				i.Logger.Errorf("Failed to index document: %s: %s", res.Error.Type, res.Error.Reason)
			}
		},
	}

	// Scripted update
	if script != "" {
		item.Action = "update"

		// Create request body
		document = &ScriptedUpdate{
			ScriptedUpsert: true,
			Script: Script{
				Source: script,
				Lang:   "painless",
				Params: document,
			},
		}
	}

	// Marshal document to JSON
	bodyJson, err := json.Marshal(document)
	if err != nil {
		return err
	}

	// Set request body
	item.Body = bytes.NewReader(bodyJson)

	// Add to bulk indexer
	return i.Client.Add(ctx, item)
}

// Stats for indexer
func (i *Indexer) Stats() esutil.BulkIndexerStats {
	return i.Client.Stats()
}

// Close and flush the indexer
func (i *Indexer) Close(ctx context.Context) error {
	return i.Client.Close(ctx)
}
