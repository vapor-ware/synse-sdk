// Synse SDK
// Copyright (c) 2017-2022 Vapor IO
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
	"bytes"
	"fmt"
)

// MultiError is a collection of errors that also fulfills the error interface.
//
// It can be used to aggregate errors and then return them all at once.
type MultiError struct {
	// Errors is a the collection of errors that the MultiError keeps track of.
	Errors []error

	// For is a string that describes the process/function that the MultiError
	// is used for. This is optional.
	For string

	// Context is a general purpose mapping that can be used to store context
	// information about errors, such as their source. This is useful when passing
	// the MultiError across different scopes where not all info may be present.
	Context map[string]string
}

// NewMultiError creates a new instance of a MultiError.
func NewMultiError(source string) *MultiError {
	return &MultiError{
		Errors:  []error{},
		For:     source,
		Context: map[string]string{},
	}
}

// Err returns the MultiError if any errors exist. Otherwise, it returns nil.
func (err *MultiError) Err() error {
	if err.HasErrors() {
		return err
	}
	return nil
}

// HasErrors checks to see if the MultiError is tracking any errors.
func (err *MultiError) HasErrors() bool {
	return len(err.Errors) != 0
}

// Add adds an error to the MultiError.
func (err *MultiError) Add(e error) {
	err.Errors = append(err.Errors, e)
}

// Error returns the error string
func (err MultiError) Error() string {
	if len(err.Errors) == 0 {
		return ""
	}

	src := err.For
	if src == "" {
		src = "unspecified"
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%d error(s) for: %s\n", len(err.Errors), src) // nolint: gas, errcheck

	for _, err := range err.Errors {
		fmt.Fprintf(&buf, "%s\n", err.Error()) // nolint: gas, errcheck
	}

	return buf.String()
}
