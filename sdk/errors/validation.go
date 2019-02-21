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

// ValidationError is used as a general error during config validation.
type ValidationError struct {
	// source is the source of the configuration which caused the validation error.
	source string

	// msg is the error message.
	msg string
}

// NewValidationError returns a new instance of a ValidationError.
func NewValidationError(source, msg string) *ValidationError {
	return &ValidationError{
		source: source,
		msg:    msg,
	}
}

// Error returns the error string and fulfils the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validating config %s: %s", e.source, e.msg)
}

// FieldRequired is an error returned when a configuration is being validated and
// a field is not filled, but it is required.
type FieldRequired struct {
	// source is the source of the configuration which caused the validation error.
	source string

	// field is the field which is not supported
	field string
}

// NewFieldRequiredError returns a new instance of a FieldRequired error.
func NewFieldRequiredError(source, field string) *FieldRequired {
	return &FieldRequired{
		source: source,
		field:  field,
	}
}

func (e *FieldRequired) Error() string {
	return fmt.Sprintf(
		"validating config %s: missing required field '%s'",
		e.source, e.field,
	)
}

// InvalidValue is an error returned when a configuration is being validated and
// a field does not contain the expected data.
type InvalidValue struct {
	// source is the source of the configuration which caused the validation error.
	source string

	// field is the field which is not supported
	field string

	// needs is a string that specifies what the field needs
	needs string
}

// NewInvalidValueError returns a new instance of an InvalidValue error.
func NewInvalidValueError(source, field, needs string) *InvalidValue {
	return &InvalidValue{
		source: source,
		field:  field,
		needs:  needs,
	}
}

func (e *InvalidValue) Error() string {
	return fmt.Sprintf(
		"validating config %s: invalid value for field '%s'. must be %s",
		e.source, e.field, e.needs,
	)
}
