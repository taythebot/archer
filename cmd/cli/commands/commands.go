package commands

import (
	"github.com/taythebot/archer/cmd/cli/runner"
	archerCli "github.com/taythebot/archer/internal/cli"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// runCommand is a wrapper to initialize the runner and run commands
func runCommand(command func(commandLine archerCli.CommandLine, runner *runner.Runner) error) func(context *cli.Context) error {
	return func(context *cli.Context) error {
		var (
			cmd   = &archerCli.ContextCommandLine{Context: context}
			debug = context.Bool("debug")
		)

		// Configure logger
		if debug {
			log.SetLevel(log.DebugLevel)
		}

		// Create new runner
		r, err := runner.New(cmd.ConfigFile(), debug)
		if err != nil {
			return err
		}

		// Run command
		if err := command(cmd, r); err != nil {
			log.Fatalf("Exiting program: %s", err)
		}

		return nil
	}
}

// Commands contains CLI commands
var Commands = []*cli.Command{
	{
		Name:  "new",
		Usage: "Create a new scan",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{Name: "module", Aliases: []string{"m"}, Usage: "module to execute", Required: true},
			&cli.StringSliceFlag{Name: "target", Aliases: []string{"t"}, Usage: "targets to scan"},
			&cli.IntSliceFlag{Name: "port", Aliases: []string{"p"}, Usage: "ports to scan", Required: true},
			&cli.StringFlag{Name: "list", Aliases: []string{"l"}, Usage: "list of targets to scan (one per line)"},
		},
		Action: runCommand(newScan),
	},
}
