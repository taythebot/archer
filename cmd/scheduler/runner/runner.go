package runner

import (
	"context"
	"errors"
	"fmt"

	"github.com/taythebot/archer/internal/yaml"
	"github.com/taythebot/archer/pkg/coordinator"
	"github.com/taythebot/archer/pkg/elasticsearch"
	"github.com/taythebot/archer/pkg/model"
	"github.com/taythebot/archer/pkg/module/httpx"
	"github.com/taythebot/archer/pkg/module/nuclei"
	"github.com/taythebot/archer/pkg/queue"
	"github.com/taythebot/archer/pkg/scheduler"
	taskHandler "github.com/taythebot/archer/pkg/task"
	"github.com/taythebot/archer/pkg/types"

	"github.com/hibiken/asynq"
	log "github.com/sirupsen/logrus"
)

// Runner for scheduler
type Runner struct {
	Config        *types.SchedulerConfig
	Coordinator   *coordinator.Client
	Elasticsearch *elasticsearch.Client
	QueueClient   *queue.Client
	QueueServer   *queue.Server
	Scheduler     *scheduler.Scheduler
}

// New creates a new Runner instance
func New(configFile string, debug bool) (*Runner, error) {
	// Create new YAML validator
	y, err := yaml.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create YAML validator: %s", err)
	}

	// Parse config file
	parsed, err := y.ValidateFile(configFile, &types.SchedulerConfig{})
	if err != nil {
		log.Error(y.FormatError(err))
		return nil, fmt.Errorf("failed to parse config file: %s", err)
	}

	// Type assertion for config
	config, ok := parsed.(*types.SchedulerConfig)
	if !ok {
		return nil, errors.New("failed to parse config file: types assertion failed")
	}

	// Create base runner
	runner := &Runner{Config: config}

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

	// Initialize Queue client
	log.Debug("Initializing Queue client")
	queueClient, err := queue.NewClient(config.RedisClient)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Queue client: %s", err)
	}
	runner.QueueClient = queueClient

	// Initialize Queue server
	log.Debug("Initializing Queue server")
	runner.QueueServer = queue.NewServer(config.RedisServer, []string{"scheduler"}, config.Concurrency, config.ID, coord, debug)

	// Connect to Database
	log.Debug("Initializing Postgresql")
	db, err := model.ConnectToDB(config.Postgresql)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Postgresql: %s", err)
	}

	// Initialize Scheduler
	log.Debug("Initializing Scheduler")
	runner.Scheduler = scheduler.New(db.DB, runner.QueueClient)

	return runner, nil
}

// Start the runner
func (r *Runner) Start() error {
	// Create new task handler
	handler := taskHandler.New(r.Elasticsearch, r.Coordinator, r.Scheduler)

	// Register handlers
	log.Debug("Registering handlers")
	if err := handler.Handle("httpx", &httpx.Module{}); err != nil {
		return fmt.Errorf("failed to register Httpx task handler: %s", err)
	}
	if err := handler.Handle("nuclei", &nuclei.Module{}); err != nil {
		return fmt.Errorf("failed to register Nuclei task handler: %s", err)
	}

	return r.QueueServer.Asynq.Run(asynq.HandlerFunc(handler.ProcessSchedulerTask))
}
