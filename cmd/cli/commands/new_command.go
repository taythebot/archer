package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/taythebot/archer/cmd/cli/runner"
	archerCli "github.com/taythebot/archer/internal/cli"
	"github.com/taythebot/archer/internal/file"
	"github.com/taythebot/archer/pkg/types"

	log "github.com/sirupsen/logrus"
)

// newScan creates a new scan
func newScan(c archerCli.CommandLine, r *runner.Runner) error {
	// Get CLI options
	modules := c.StringSlice("module")
	targets := c.StringSlice("target")
	ports := c.IntSlice("port")
	list := c.String("list")

	// Add all modules
	if len(modules) == 1 && modules[0] == "all" {
		modules = types.Modules
	} else {
		// Validate modules
		for _, m := range modules {
			var valid bool
			for _, module := range types.Modules {
				if module == m {
					valid = true
					break
				}
			}

			if !valid {
				return fmt.Errorf("module must be one of %s, all", strings.Join(types.Modules, ", "))
			}
		}
	}

	// Read targets from list
	if len(targets) == 0 {
		if list == "" {
			return errors.New("no targets provided")
		}

		// Read file
		t, err := file.ReadFile(list)
		if err != nil {
			return fmt.Errorf("failed to read targets list: %s", err)
		} else if len(t) == 0 {
			return errors.New("no targets found in list")
		}

		targets = t
	}

	// Convert ports to uint16
	var p []uint16
	for _, port := range ports {
		p = append(p, uint16(port))
	}

	// Create new scan
	log.Info("Creating new scan")
	scan, err := r.Coordinator.NewScan(context.Background(), targets, p, modules)
	if err != nil {
		return fmt.Errorf("failed to create new scan: %s", err)
	}

	log.Infof("Successfully created scan %s", scan.ID)
	return nil
}
