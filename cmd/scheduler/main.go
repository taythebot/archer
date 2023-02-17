package main

import (
	"flag"

	"github.com/taythebot/archer/cmd/scheduler/runner"

	log "github.com/sirupsen/logrus"
)

func main() {
	// Parse flags
	configFile := flag.String("config", "configs/scheduler.yaml", "Config file")
	debug := flag.Bool("debug", false, "Enable debug level logging")
	flag.Parse()

	// Validate flags
	if *configFile == "" {
		log.Fatal("Exiting Program: config file is required")
	}

	// Set log level
	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	// Create runner
	r, err := runner.New(*configFile, *debug)
	if err != nil {
		log.Fatalf("Exiting Program: %s", err)
	}

	// Start runner
	log.Infof("Starting worker with %d threads", r.Config.Concurrency)
	if err := r.Start(); err != nil {
		log.Fatalf("Exiting Program: %s", err)
	}
}
