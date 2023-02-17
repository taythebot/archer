package cli

import "github.com/urfave/cli/v2"

/*
	This code was inspired by Grafana's CLI service
	All credits go to Grafana
*/

type CommandLine interface {
	ShowHelp() error
	ShowVersion()
	Application() *cli.App
	Args() cli.Args
	Bool(name string) bool
	Int(name string) int
	IntSlice(name string) []int
	String(name string) string
	StringSlice(name string) []string
	FlagNames() (names []string)
	Generic(name string) interface{}
}

type ContextCommandLine struct {
	*cli.Context
}

func (c *ContextCommandLine) ShowHelp() error {
	return cli.ShowCommandHelp(c.Context, c.Command.Name)
}

func (c *ContextCommandLine) ShowVersion() {
	cli.ShowVersion(c.Context)
}

func (c *ContextCommandLine) Application() *cli.App {
	return c.App
}

func (c *ContextCommandLine) ConfigFile() string {
	return c.String("config")
}
