package runner

import (
	"context"
	"errors"
	"fmt"

	"github.com/taythebot/archer/internal/yaml"
	"github.com/taythebot/archer/pkg/coordinator"
	"github.com/taythebot/archer/pkg/types"

	log "github.com/sirupsen/logrus"
)

// Runner ...
type Runner struct {
	Config      *types.CliConfig
	Coordinator *coordinator.Client
}

// New creates a new Runner instance
func New(configFile string, debug bool) (*Runner, error) {
	// Create new YAML validator
	y, err := yaml.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create YAML validator: %s", err)
	}

	// Parse config file
	log.Debug("Parsing configuration file")
	parsed, err := y.ValidateFile(configFile, &types.CliConfig{})
	if err != nil {
		log.Error(y.FormatError(err))
		return nil, fmt.Errorf("failed to parse config file: %s", err)
	}
	config, ok := parsed.(*types.CliConfig)
	if !ok {
		return nil, errors.New("failed to parse config file: types assertion failed")
	}

	// Initialize Coordinator
	log.Debug("Initializing Coordinator")
	coord := coordinator.New(config.Coordinator, "", debug)
	if _, err := coord.Health(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize Coordinator: %s", err)
	}

	return &Runner{
		Config:      config,
		Coordinator: coord,
	}, nil
}
