package nuclei

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
	hostRegexp = regexp.MustCompile(`([\d.]+):(\d+)`)

	// elasticScript is the painless script for updating documents
	elasticScript = `
		// Add detection
		if (ctx._source.detections == null) {
			ctx._source.detections = new ArrayList();
			ctx._source.detections.add(params.detection);	
		} else {
			// Delete existing entry
			def push = true;
			for (int i = 0; i < ctx._source.detections.length; i++) {
				HashMap curr = ctx._source.detections[i];
				if (curr.port == params.detection.port && curr.template_id == params.detection.template_id) {
					push = false;
					break;
				}
			}

			// Push into array
			if (push) {
				ctx._source.detections.add(params.detection);
			}
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

	// Build command arguments
	args := []string{
		"-l=" + configFile,
		"-timeout=" + strconv.Itoa(int(m.Config.Timeout)),
		"-retries=" + strconv.Itoa(int(m.Config.Retries)),
		"-retries=" + strconv.Itoa(int(m.Config.RateLimit)),
		"-bs=" + strconv.Itoa(int(m.Config.BulkSize)),
		"-c=" + strconv.Itoa(int(m.Config.Concurrency)),
		"-stats",
		"-json",
	}

	// Add template types
	for _, t := range modulePayload.TemplateTypes {
		args = append(args, "-type="+t)
	}

	// Add proxies
	for _, proxy := range m.Config.Proxies {
		args = append(args, "-proxy="+proxy)
	}

	// Create command
	cmd := exec.New(m.Config.Binary, args...)

	// Create new results channel
	results := make(chan types.TaskResult, 1)

	// Create new sync error group
	g, gCtx := errgroup.WithContext(ctx)

	// Start Httpx
	if err = cmd.Start(ctx); err != nil {
		return 0, fmt.Errorf("failed to execute Httpx: %s", err)
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

	// Wait for Nuclei to finish
	g.Go(func() error {
		if err := cmd.Wait(); err != nil {
			return fmt.Errorf("failed to execute Httpx: %s", err)
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

	// Remove config files
	if !m.Config.PersistConfig {
		logger.Info("Removing config files")
		if err := os.Remove(configFile); err != nil {
			logger.Errorf("Failed to remove config file '%s': %s", configFile, err)
		}
	}

	return
}

// ProcessStdout parses the stdout from the Httpx process
func (m *Module) ProcessStdout(ctx context.Context, stdout *bufio.Scanner, results chan<- types.TaskResult, scanId, taskId string, log *logrus.Entry) error {
	for stdout.Scan() {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Unmarshal payload
		parsed := &Output{}
		if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
			log.Errorf("Failed to parse output: %s", err)
			continue
		}

		// Parse matched at
		matches := hostRegexp.FindStringSubmatch(parsed.MatchedAt)
		if len(matches) != 3 {
			log.Errorf("Failed to parse host '%s'", parsed.MatchedAt)
			continue
		}

		// Convert port to int
		port, err := strconv.Atoi(matches[2])
		if err != nil {
			log.Errorf("Failed to convert port '%s' to int", matches[2])
			continue
		}

		// Create detection
		detection := ResultDetection{
			Port:             port,
			TemplateId:       parsed.TemplateId,
			Type:             parsed.Type,
			ExtractedResults: parsed.ExtractedResults,
			MatcherName:      parsed.MatcherName,
			MatchedAt:        parsed.MatchedAt,
			Name:             parsed.Info.Name,
			Description:      parsed.Info.Description,
			Severity:         parsed.Info.Severity,
			Tags:             parsed.Info.Tags,
			Metadata: ResultDetectionMetadata{
				Module:    "nuclei",
				Task:      taskId,
				Timestamp: time.Now().UTC(),
			},
		}

		// Create document ID
		h := sha1.New()
		h.Write([]byte(matches[1] + scanId))
		id := hex.EncodeToString(h.Sum(nil))

		// Send to results channel
		results <- types.TaskResult{
			ID: id,
			Doc: Result{
				IP:        matches[1],
				Detection: detection,
				Scan:      scanId,
				Timestamp: detection.Metadata.Timestamp,
			},
		}

		// Output
		log.Infof("Found %s [%s] at %s", parsed.TemplateId, parsed.Info.Severity, parsed.MatchedAt)
	}

	return nil
}

// ProcessStderr processes the stderr output from the Httpx process
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
