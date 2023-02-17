package httpx

import (
	"context"
	"fmt"
	"strconv"

	"github.com/taythebot/archer/pkg/elasticsearch"
	"github.com/taythebot/archer/pkg/module/masscan"
	"github.com/taythebot/archer/pkg/types"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var (
	// Source is a list of fields to grab from Elasticsearch
	source = []string{"ip", "ports.port", "timestamp"}

	// maxTargets per batch
	maxTargets = 50000
)

// Schedule new tasks
func (m *Module) Schedule(ctx context.Context, payload types.SchedulerPayload, scheduler types.Scheduler, es *elasticsearch.Client, logger *logrus.Entry) (total uint32, err error) {
	// Store targets in memory
	var targets []string

	// Create results channel
	results := make(chan string, 1)

	// Create sync error group
	g, ctx := errgroup.WithContext(ctx)

	// Process results
	g.Go(func() error {
		// Keep track of targets in memory
		var i int

		// Consume results
		for result := range results {
			// Check for context cancellation
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			// Append to slice
			targets = append(targets, result)

			// Increase counter
			i++

			// Schedule if max target
			if i == maxTargets {
				if _, err := scheduler.Httpx(ctx, payload.Scan, targets); err != nil {
					return fmt.Errorf("failed to schedule tasks: %s", err)
				}

				// Increase total
				total++

				// Reset
				targets = []string{}
				i = 0
			}
		}

		// Schedule remaining before exit
		if len(targets) > 0 {
			if _, err := scheduler.Httpx(ctx, payload.Scan, targets); err != nil {
				return fmt.Errorf("failed to schedule remaining tasks: %s", err)
			}

			// Increase total
			total++
		}

		return nil
	})

	// Fetch results
	g.Go(func() error {
		defer close(results)

		// Track total results
		var totalTargets int

		// Construct query
		query := "scan:" + payload.Scan + " AND ports.metadata.task:" + payload.Task

		// Create new PIT
		pit, err := es.OpenPit(ctx, "1m")
		if err != nil {
			return fmt.Errorf("failed to create PIT: %s", err)
		}

		// Loop search until all results are fetched
		var searchAfter []interface{}
		for {
			// Check for context cancellation
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			// Get results
			result, err := es.SearchAfter(ctx, query, 100, source, searchAfter, "timestamp:desc", pit, "1m")
			if err != nil {
				return fmt.Errorf("failed to search Elasticsearch: %s", err)
			}

			// Get total results
			totalResults := len(result.Hits.Hits) - 1

			// Exit loop if no results
			if totalResults == -1 {
				break
			}

			logger.Infof("Found %d results", totalResults+1)

			// Parse results
			for index, hit := range result.Hits.Hits {
				// Unmarshal document
				var doc masscan.Result
				if err := json.Unmarshal(hit.Source, &doc); err != nil {
					logger.Errorf("Failed to unmarshal document '%s': %s", hit.ID, err)
					continue
				}

				// Create targets
				for _, port := range doc.Ports {
					// Send target
					results <- doc.IP + ":" + strconv.Itoa(port.Port)

					// Increase counter
					totalTargets++
				}

				// Set search after
				if index == totalResults {
					searchAfter = hit.Sort
				}
			}
		}

		logger.Infof("Total of %d targets found", totalTargets)
		return nil
	})

	// Wait for workers
	if err := g.Wait(); err != nil {
		return total, err
	}

	return
}
