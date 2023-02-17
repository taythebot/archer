package cli

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func Run(commands []*cli.Command, action func(context *cli.Context) error) {
	app := &cli.App{
		Name:    "Archer",
		Usage:   "Distributed scanner for the masses",
		Version: Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Aliases:  []string{"c"},
				Usage:    "Path to configuration file",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "Enable debug level logging",
			},
		},
		Commands:        commands,
		CommandNotFound: cmdNotFound,
	}

	// Default action
	app.Action = action

	// Display banner
	app.Before = func(_ *cli.Context) error {
		ShowBanner()
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Program Exiting: %s", err)
	}
}

// cmdNotFound handles errors for when a command is not found
func cmdNotFound(_ *cli.Context, command string) {
	log.Fatalf("Program Exiting: Unknown command '%s'. Try --help to see commands.", command)
}
