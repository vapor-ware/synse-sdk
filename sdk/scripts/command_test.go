// Synse SDK
// Copyright (c) 2019 Vapor IO
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
