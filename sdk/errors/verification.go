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
	"fmt"
)

// VerificationConflict is an error that is used when the config verification
// process detects that a config is invalid because there are some components
// in conflict with one another.
type VerificationConflict struct {
	// configType is the type of config that is being verified.
	configType string

	// msg is the error message.
	msg string
}

// NewVerificationConflictError creates a new instance of the VerificationConflict error.
func NewVerificationConflictError(configType, msg string) *VerificationConflict {
	return &VerificationConflict{
		configType: configType,
		msg:        msg,
	}
}

// Error returns the error string and fulfils the error interface.
func (e *VerificationConflict) Error() string {
	return fmt.Sprintf("conflict detected when verifying %s config: %s", e.configType, e.msg)
}

// VerificationInvalid is an error that is used when the config verification
// process detects that a config is invalid because a piece of data being referenced
// is expected but not found.
type VerificationInvalid struct {
	// configType is the type of config that is being verified.
	configType string

	// msg is the error message.
	msg string
}

// NewVerificationInvalidError creates a new instance of the VerificationInvalid error.
func NewVerificationInvalidError(configType, msg string) *VerificationInvalid {
	return &VerificationInvalid{
		configType: configType,
		msg:        msg,
	}
}

// Error returns the error string and fulfils the error interface.
func (e *VerificationInvalid) Error() string {
	return fmt.Sprintf("%s config invalid: %s", e.configType, e.msg)
}
