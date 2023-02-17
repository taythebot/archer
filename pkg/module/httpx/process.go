package httpx

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
	"strings"
	"time"

	"github.com/taythebot/archer/internal/exec"
	"github.com/taythebot/archer/pkg/elasticsearch"
	"github.com/taythebot/archer/pkg/types"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var (
	// httpHeaderRegex is used to parse HTTP headers
	httpHeaderRegex = regexp.MustCompile(`(?P<Key>[\w-]+):\s(?P<Value>.+)`)

	// elasticScript is the painless script for updating documents
	elasticScript = `
		// Add http
		if (ctx._source.http == null) {
			ctx._source.http = new ArrayList();
			ctx._source.http.add(params.output);
		} else {
			// Delete existing entry
			def push = true;
			for (int i = 0; i < ctx._source.http.length; i++) {
				HashMap curr = ctx._source.http[i];
				if (curr.port == params.output.port) {
					// Check if existing entry is HTTPS
					if (curr.scheme == 'HTTPS' && params.output.scheme == 'HTTP') {
						push = false;
					} else {
						ctx._source.http.remove(i);
					}

					break;
				}
			}

			// Push into array
			if (push) {
				ctx._source.http.add(params.output);
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
		"-follow-redirects",
		"-random-agent",
		"-status-code",
		"-server",
		"-tech-detect",
		"-tls-grab",
		"-title",
		"-no-fallback",
		"-include-chain",
		"-include-response",
		"-json",
	}

	// Add proxies
	if m.Config.HttpProxy != "" {
		args = append(args, "-http-proxy="+m.Config.HttpProxy)
	} else if m.Config.SocksProxy != "" {
		args = append(args, "-socks-proxy="+m.Config.SocksProxy)
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

	// Wait for Httpx to finish
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

		// Parse ip out of input
		inputMatches := strings.Split(parsed.Input, ":")
		if len(inputMatches) != 2 {
			log.Errorf("Failed to parse IP out of input '%s'", parsed.Input)
			continue
		}
		ip := inputMatches[0]

		// Convert port to int
		port, err := strconv.Atoi(parsed.Port)
		if err != nil {
			log.Errorf("Failed to convert port '%s' to int: %s", parsed.Port, err)
			continue
		}

		// Create base http result
		http := ResultHttp{
			Port: port,
			Body: parsed.Body,
			Hashes: ResultHttpHashes{
				BodyMmh3:     parsed.Hash.BodyMmh3,
				BodySha256:   parsed.Hash.BodySha256,
				HeaderMmh3:   parsed.Hash.HeaderMmh3,
				HeaderSha256: parsed.Hash.HeaderSha256,
			},
			Technologies: parsed.Tech,
			Title:        parsed.Title,
			Scheme:       parsed.Scheme,
			StatusCode:   parsed.StatusCode,
			Metadata: ResultHttpMetadata{
				Module:    "httpx",
				Task:      taskId,
				Timestamp: time.Now().UTC(),
			},
		}

		// Add CSP
		if parsed.Csp.Domains != nil {
			http.Csp = parsed.Csp.Domains
		}

		// Add Tls
		if parsed.TLS.Version != "" {
			http.Tls = parsed.TLS
		}

		// Parse and add headers
		headers := make(map[string]string)
		for _, v := range strings.Split(parsed.RawHeader, "\r\n") {
			matches := httpHeaderRegex.FindStringSubmatch(v)
			if len(matches) == 3 {
				headers[strings.ToLower(matches[1])] = matches[2]
			}
		}
		http.Headers = headers

		// Add final url
		if parsed.FinalURL != "" {
			http.Redirects = ResultHttpRedirects{
				FinalUrl: parsed.FinalURL,
			}
		}

		// Add redirects
		if len(parsed.Chain) > 0 {
			for _, chain := range parsed.Chain {
				http.Redirects.Chains = append(http.Redirects.Chains, ResultHttpRedirectsChain{
					Request:    chain.Request,
					Response:   chain.Response,
					StatusCode: chain.StatusCode,
					Location:   chain.Location,
					RequestUrl: chain.RequestURL,
				})
			}
		}

		// Create document ID
		h := sha1.New()
		h.Write([]byte(ip + scanId))
		id := hex.EncodeToString(h.Sum(nil))

		// Send to results channel
		results <- types.TaskResult{
			ID: id,
			Doc: Result{
				IP:        ip,
				Output:    http,
				Scan:      scanId,
				Timestamp: http.Metadata.Timestamp,
			},
		}

		// Output
		log.Infof("Found %s service with status code %d at %s", strings.ToUpper(parsed.Scheme), parsed.StatusCode, parsed.Input)
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
