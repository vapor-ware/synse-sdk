package scripts

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewCommand gets a new instance of a Command with args.
func TestNewCommand(t *testing.T) {
	command := NewCommand("echo", "Hello World", "\t", "foo bar")

	assert.IsType(t, &Command{}, command)
	assert.Equal(t, "echo", command.bin)
	assert.Equal(t, 0, command.stdout.Len())
	assert.Equal(t, 0, command.stderr.Len())
}

// TestNewCommand2 gets a new instance of a Command with no args.
func TestNewCommand2(t *testing.T) {
	command := NewCommand("echo")

	assert.IsType(t, &Command{}, command)
	assert.Equal(t, "echo", command.bin)
	assert.Equal(t, 0, command.stdout.Len())
	assert.Equal(t, 0, command.stderr.Len())
}

// TestNewCommand3 gets a new instance of a Command with no binary and no args.
func TestNewCommand3(t *testing.T) {
	command := NewCommand("")

	assert.IsType(t, &Command{}, command)
	assert.Equal(t, "", command.bin)
	assert.Equal(t, 0, command.stdout.Len())
	assert.Equal(t, 0, command.stderr.Len())
}

// TestCommand_Run runs a command successfully.
func TestCommand_Run(t *testing.T) {
	command := NewCommand("echo", "Hello World")
	err := command.Run()

	assert.NoError(t, err)
	assert.Equal(t, "Hello World\n", command.Stdout())
	assert.Equal(t, "", command.Stderr())
}

// TestCommand_Run2 runs a command unsuccessfully.
func TestCommand_Run2(t *testing.T) {
	command := NewCommand("this-is-not-a-bin", "Hello World")
	err := command.Run()

	assert.Error(t, err)
	assert.Equal(t, "command unable to find binary: this-is-not-a-bin", err.Error())
	assert.Equal(t, "", command.Stdout())
	assert.Equal(t, "", command.Stderr())
}

// TestCommand_Run3 runs a command from a file, not from path. This file
// will exist, but since it is not executable, it will fail to run.
func TestCommand_Run3(t *testing.T) {
	// Create the temporary file
	tmpfile, err := ioutil.TempFile("", "foobar")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.Remove(tmpfile.Name())
		if err != nil {
			t.Error(err)
		}
	}()

	command := NewCommand(tmpfile.Name(), "some", "args")
	err = command.Run()

	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "permission denied"))
	assert.Equal(t, "", command.Stdout())
	assert.Equal(t, "", command.Stderr())
}

// TestCommand_Run4 runs a command with a bin that exists and is executable, but
// it fails. This should output to stderr.
func TestCommand_Run4(t *testing.T) {
	command := NewCommand("ls", "abcdefghijklmnopqrstuvwxyz")
	err := command.Run()

	assert.Error(t, err)
	assert.Equal(t, "", command.Stdout())
	assert.True(t, strings.Contains(command.Stderr(), "No such file or directory"))
}
