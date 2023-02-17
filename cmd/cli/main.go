package main

import (
	"github.com/taythebot/archer/cmd/cli/commands"
	archerCli "github.com/taythebot/archer/internal/cli"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	archerCli.Run(commands.Commands, func(c *cli.Context) error {
		log.Fatal("Exiting Program: No command given. Try --help to see commands.")
		return nil
	})
}
