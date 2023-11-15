package runner

import (
    "context"
    "errors"
    "fmt"

    "github.com/taythebot/archer/internal/yaml"
    "github.com/taythebot/archer/pkg/coordinator"
    "github.com/taythebot/archer/pkg/elasticsearch"
    "github.com/taythebot/archer/pkg/module/httpx"
    "github.com/taythebot/archer/pkg/module/masscan"
    "github.com/taythebot/archer/pkg/module/nuclei"
    "github.com/taythebot/archer/pkg/queue"
    taskHandler "github.com/taythebot/archer/pkg/task"
    "github.com/taythebot/archer/pkg/types"

    "github.com/hibiken/asynq"
    log "github.com/sirupsen/logrus"
)

// Runner for worker
type Runner struct {
    Config        *types.WorkerConfig
    Coordinator   *coordinator.Client
    Elasticsearch *elasticsearch.Client
    Queue         *queue.Server
    Masscan       *masscan.Module
    Httpx         *httpx.Module
    Nuclei        *nuclei.Module
}

// New creates a new Runner instance
func New(configFile string, debug bool) (*Runner, error) {
    config := defaultConfig()

    // Create new YAML validator
    y, err := yaml.New()
    if err != nil {
        return nil, fmt.Errorf("failed to create YAML validator: %s", err)
    }

    // Parse config file
    parsed, err := y.ValidateFile(configFile, config)
    if err != nil {
        log.Error(y.FormatError(err))
        return nil, fmt.Errorf("failed to parse config file: %s", err)
    }

    // Type assertion for config
    config, ok := parsed.(*types.WorkerConfig)
    if !ok {
        return nil, errors.New("failed to parse config file: types assertion failed")
    }

    // Create base runner
    runner := &Runner{Config: config}

    // Validate modules
    for _, m := range config.Modules {
        var valid bool
        for _, moduleName := range types.Modules {
            if moduleName == m {
                valid = true
                break
            }
        }

        // Check if valid
        if !valid {
            return nil, fmt.Errorf("invalid module '%s' provided", m)
        }

        // Custom warning
        if m == "masscan" && config.Concurrency > 1 {
            log.Warn("Using more than 1 concurrency for Masscan is not recommended!")
        }

 		// Initialize module
         switch m {
		case "masscan":
			if config.Masscan == nil {
				return nil, errors.New("failed to initialize Masscan: config not found")
			}

			runner.Masscan = masscan.New(*config.Masscan)
		case "httpx":
			if config.Httpx == nil {
				return nil, errors.New("failed to initialize Httpx: config not found")
			}

			if config.Httpx.HttpProxy != "" && config.Httpx.SocksProxy != "" {
				return nil, errors.New("failed to initialize Httpx: cannot use HTTP and Socks proxy at the same time")
			}

			runner.Httpx = httpx.New(*config.Httpx)
		case "nuclei":
			if config.Nuclei == nil {
				return nil, errors.New("failed to iniitalize Nuclei: config not found")
			}

			runner.Nuclei = nuclei.New(*config.Nuclei)
		default:
			return nil, fmt.Errorf("module '%s' not found", m)
		}
	}

	// Initialize Coordinator
	log.Debug("Initializing Coordinator")
	coord := coordinator.New(config.Coordinator, config.ID, debug)
	if _, err := coord.Health(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize Coordinator: %s", err)
	}
	runner.Coordinator = coord

	// Initialize Elasticsearch
	log.Debug("Initializing Elasticsearch")
	es, err := elasticsearch.New(config.Elasticsearch, debug)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Elasticsearch: %s", err)
	}

	runner.Elasticsearch = es

	// Initialize Queue server
	log.Debug("Initializing Queue server")
	runner.Queue = queue.NewServer(config.Redis, config.Modules, config.Concurrency, config.ID, coord, debug)

	return runner, nil
}

// Start the runner
func (r *Runner) Start() error {
	// Create new task handler
	handler := taskHandler.New(r.Elasticsearch, r.Coordinator, nil)

	// Register handlers
	log.Debug("Registering handlers")
	for _, module := range r.Config.Modules {
		var err error

		log.Debugf("Registering handler for module '%s'", module)

		switch module {
		case "masscan":
			err = handler.Handle("masscan", r.Masscan)
		case "httpx":
			err = handler.Handle("httpx", r.Httpx)
		case "nuclei":
			err = handler.Handle("nuclei", r.Nuclei)
		default:
			err = fmt.Errorf("failed to find task handler for module '%s'", module)
		}

		if err != nil {
			return err
		}
	}

	return r.Queue.Asynq.Run(asynq.HandlerFunc(handler.ProcessScanTask))
}
