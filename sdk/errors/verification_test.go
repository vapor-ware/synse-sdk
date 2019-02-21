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

func TestNewVerificationConflictError(t *testing.T) {
	err := NewVerificationConflictError("test", "message")

	assert.IsType(t, &VerificationConflict{}, err)
	assert.Equal(t, "test", err.configType)
	assert.Equal(t, "message", err.msg)
}

func TestVerificationConflict_Error(t *testing.T) {
	err := NewVerificationConflictError("test", "message")
	out := err.Error()

	assert.Equal(t, "conflict detected when verifying test config: message", out)
}

func TestNewVerificationInvalidError(t *testing.T) {
	err := NewVerificationInvalidError("test", "message")

	assert.IsType(t, &VerificationInvalid{}, err)
	assert.Equal(t, "test", err.configType)
	assert.Equal(t, "message", err.msg)
}

func TestVerificationInvalid_Error(t *testing.T) {
	err := NewVerificationInvalidError("test", "message")
	out := err.Error()

	assert.Equal(t, "test config invalid: message", out)
}
