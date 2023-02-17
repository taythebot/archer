package masscan

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/taythebot/archer/internal/exec"
	"github.com/taythebot/archer/pkg/elasticsearch"
	"github.com/taythebot/archer/pkg/types"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var (
	// regularRe is the regex for regular Masscan outputs
	regularRe = regexp.MustCompile(`Discovered\sopen\sport\s(\d+)/(tcp|udp)\son\s([\w.]+)`)

	// elasticScript is the painless script for updating documents
	elasticScript = `
		// Add port
		if (ctx._source.ports == null) {
			ctx._source.ports = new ArrayList();
			ctx._source.ports.add(params.port);
		} else if (!ctx._source.ports.contains(params.port)) {
			ctx._source.ports.add(params.port);    
		}
		
		// Add IP and metadata
		ctx._source.ip = params.ip;
		ctx._source.scan = params.scan;
		ctx._source.timestamp = params.timestamp;
	`
)

// ProcessTask processes a task from the queue
func (m *Module) ProcessTask(ctx context.Context, taskId string, payload types.TaskPayload, indexer *elasticsearch.Indexer, logger *logrus.Entry) (total uint32, err error) {
	// Type assertion to local payload
	modulePayload, ok := payload.(*Payload)
	if !ok {
		return 0, errors.New("failed to perform type assertion to local payload")
	}

	// Build config
	logger.Info("Building config file")
	configFile, err := m.BuildConfig(taskId, modulePayload)
	if err != nil {
		return 0, fmt.Errorf("failed to build config file: %s", err)
	}
	logger.Infof("Created config file %s", configFile)

	// Create exec args
	args := []string{"-c", configFile}
	if m.Config.ExcludeFile != "" {
		args = append(args, "--excludefile", m.Config.ExcludeFile)
	}

	// Create new exec
	cmd := exec.New(m.Config.Binary, args...)

	// Create new results channel
	results := make(chan types.TaskResult, 1)

	// Create new sync error group
	g, gCtx := errgroup.WithContext(ctx)

	// Start Masscan
	if err = cmd.Start(ctx); err != nil {
		return 0, fmt.Errorf("failed to execute Masscan: %s", err)
	}

	// Process results
	g.Go(func() error {
		for result := range results {
			// Check for context cancellation
			select {
			case <-gCtx.Done():
				return fmt.Errorf("failed to process results: %s", ctx.Err())
			default:
			}

			// Add to Elasticsearch indexer
			if err := indexer.Add(gCtx, result.ID, result.Doc, elasticScript); err != nil {
				return fmt.Errorf("failed to index document: %s", err)
			}

			// Increase counter
			total++
		}

		return nil
	})

	// Process outputs
	g.Go(func() error {
		if err := m.ProcessStdout(gCtx, cmd.Stdout, results, modulePayload.Scan, taskId, logger); err != nil {
			return fmt.Errorf("failed to process stdout: %s", err)
		}

		return nil
	})
	g.Go(func() error {
		if err := m.ProcessStderr(gCtx, cmd.Stderr, logger); err != nil {
			return fmt.Errorf("failed to process stderr: %s", err)
		}

		return nil
	})

	// Wait for Masscan to finish
	g.Go(func() error {
		if err := cmd.Wait(); err != nil {
			return fmt.Errorf("failed to execute Masscan: %s", err)
		}

		logger.Info("Closing results channel")
		close(results)

		return nil
	})

	// Wait for groups
	logger.Info("Executing task")
	if err := g.Wait(); err != nil {
		return total, err
	}

	// Remove config file
	if !m.Config.PersistConfig {
		logger.Info("Removing config file")
		if err := os.Remove(configFile); err != nil {
			logger.Errorf("Failed to remove config file '%s': %s", configFile, err)
		}
	}

	return
}

// ProcessStdout parses the stdout from the Masscan process
func (m *Module) ProcessStdout(ctx context.Context, stdout *bufio.Scanner, results chan<- types.TaskResult, scanId, taskId string, log *logrus.Entry) error {
	for stdout.Scan() {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Get text
		text := stdout.Text()

		// Create base document
		result := Result{
			Scan:      scanId,
			Timestamp: time.Now().UTC(),
		}

		// Parse regular output without banner Ex: Discovered open port <port>/<proto> on <ip>
		rm := regularRe.FindStringSubmatch(text)
		if len(rm) == 4 {
			// Convert port string to int
			port, err := strconv.Atoi(rm[1])
			if err != nil {
				log.Errorf("Failed to parse output: %s", err)
				continue
			}

			// Add to document
			result.Port = ResultPort{
				Port: port,
				Metadata: ResultMetadata{
					Module:    "masscan",
					Task:      taskId,
					Timestamp: result.Timestamp,
				},
			}
			result.IP = rm[3]
		}

		// Create document ID
		h := sha1.New()
		h.Write([]byte(result.IP + scanId))
		id := hex.EncodeToString(h.Sum(nil))

		// Send to results channel
		results <- types.TaskResult{ID: id, Doc: result}

		// Output
		log.Infof("Found port %d at %s", result.Port.Port, result.IP)
	}

	return nil
}

// ProcessStderr processes the stderr output from the Masscan process
func (m *Module) ProcessStderr(ctx context.Context, stderr *bufio.Scanner, log *logrus.Entry) error {
	for stderr.Scan() {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			log.Info(stderr.Text())
		}
	}

	return nil
}
