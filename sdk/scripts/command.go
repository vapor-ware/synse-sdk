package scripts

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// Command models a command to be run. This is used to run scripts
// from within a plugin.
type Command struct {
	bin string
	cmd *exec.Cmd

	stdout *bytes.Buffer
	stderr *bytes.Buffer
}

// NewCommand creates a new instance of a Command.
func NewCommand(bin string, args ...string) *Command {
	return &Command{
		bin:    bin,
		cmd:    exec.Command(bin, args...), // nolint: gas
		stdout: new(bytes.Buffer),
		stderr: new(bytes.Buffer),
	}
}

// binExists checks that the binary for the command exists on the
// local system.
func (command *Command) binExists() bool {
	// The 'bin' here could be a file (e.g. run.sh) or it could be
	// the fully qualified path to an executable (/usr/bin/foo).
	// First, check if command.bin exists as a file however it is defined.
	_, err := os.Stat(command.bin)
	if err == nil {
		return true
	}

	// If it was not found for any reason above, check to see if it
	// is on the PATH.
	_, err = exec.LookPath(command.bin)
	return err == nil
}

// Stdout returns the Command's string output to stdout.
func (command *Command) Stdout() string {
	return command.stdout.String()
}

// Stderr returns the Command's string output to stderr.
func (command *Command) Stderr() string {
	return command.stderr.String()
}

// Run executes the command.
func (command *Command) Run() error {
	binExists := command.binExists()
	if !binExists {
		return fmt.Errorf("command unable to find binary: %s", command.bin)
	}

	command.cmd.Stdout = bufio.NewWriter(command.stdout)
	command.cmd.Stderr = bufio.NewWriter(command.stderr)

	return command.cmd.Run()
}
