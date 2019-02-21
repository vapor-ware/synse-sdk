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

package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("test", "message")

	assert.IsType(t, &ValidationError{}, err)
	assert.Equal(t, "test", err.source)
	assert.Equal(t, "message", err.msg)
}

func TestValidationError_Error(t *testing.T) {
	err := NewValidationError("test", "message")
	out := err.Error()

	assert.Equal(t, "validating config test: message", out)
}

func TestNewFieldRequiredError(t *testing.T) {
	err := NewFieldRequiredError("test", "foo")

	assert.IsType(t, &FieldRequired{}, err)
	assert.Equal(t, "test", err.source)
	assert.Equal(t, "foo", err.field)
}

func TestFieldRequired_Error(t *testing.T) {
	err := NewFieldRequiredError("test", "foo")
	out := err.Error()

	assert.Equal(t, "validating config test: missing required field 'foo'", out)
}

func TestNewInvalidValueError(t *testing.T) {
	err := NewInvalidValueError("test", "foo", "greater than 2")

	assert.IsType(t, &InvalidValue{}, err)
	assert.Equal(t, "test", err.source)
	assert.Equal(t, "foo", err.field)
	assert.Equal(t, "greater than 2", err.needs)
}

func TestInvalidValue_Error(t *testing.T) {
	err := NewInvalidValueError("test", "foo", "greater than 2")
	out := err.Error()

	assert.Equal(t, "validating config test: invalid value for field 'foo'. must be greater than 2", out)
}
