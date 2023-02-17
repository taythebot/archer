package exec

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
)

// Client for executing commands
type Client struct {
	Name   string         // Name contains the main command
	Args   []string       // Args contains the command arguments
	Pid    int            // Pid contains the process id
	Cmd    *exec.Cmd      // Cmd contains the OS cmd instance
	Stderr *bufio.Scanner // Stderr contains the stderr output
	Stdout *bufio.Scanner // Stdout contains the stdout output
}

// New creates a new instance of client
func New(name string, args ...string) *Client {
	return &Client{
		Name: name,
		Args: args,
	}
}

// Start executes the command
func (c *Client) Start(ctx context.Context) error {
	// Create new command
	cmd := exec.CommandContext(ctx, c.Name, c.Args...)

	// Get outputs
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("unable to get stderr from command: %s", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("unable to get stdout from command: %s", err)
	}

	// Create outputs
	c.Stdout = bufio.NewScanner(stdout)
	c.Stderr = bufio.NewScanner(stderr)

	// Start command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to execute command: %s", err)
	}

	// Get process id
	c.Pid = cmd.Process.Pid

	// Set command
	c.Cmd = cmd

	return nil
}

// Wait for command to finish executing
func (c *Client) Wait() error {
	return c.Cmd.Wait()
}
