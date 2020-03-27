// Synse SDK
// Copyright (c) 2017-2020 Vapor IO
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

package errors

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInvalidArgumentErr tests constructing a new InvalidArgument error.
func TestInvalidArgumentErr(t *testing.T) {
	errString := "test error"
	err := InvalidArgumentErr(errString)

	assert.True(t, strings.Contains(err.Error(), "InvalidArgument"))
	assert.True(t, strings.Contains(err.Error(), errString))
}

// TestNotFoundErr tests constructing a new NotFound error.
func TestNotFoundErr(t *testing.T) {
	errString := "test error"
	err := NotFoundErr(errString)

	assert.True(t, strings.Contains(err.Error(), "NotFound"))
	assert.True(t, strings.Contains(err.Error(), errString))
}

// TestUnsupportedCommandErrorErr tests constructing and stringify-ing
// an UnsupportedCommandError error.
func TestUnsupportedCommandErrorErr(t *testing.T) {
	err := UnsupportedCommandError{}
	assert.Error(t, &err)

	assert.Equal(
		t,
		"Command not supported for given device.",
		err.Error(),
	)
}
