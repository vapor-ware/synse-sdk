package errors

import (
	"bytes"
	"fmt"
)

/*
TODO:
----------------
- add error types for validation errors
*/


// MultiError is a collection of errors that also fulfills the error interface.
//
// It can be used to aggregate errors and then return them all at once.
type MultiError struct {
	// Errors is a the collection of errors that the MultiError keeps track of.
	Errors []error

	// For is a string that describes the process/function that the MultiError
	// is used for. This is optional.
	For string
}

// NewMultiError creates a new instance of a MultiError.
func NewMultiError(source string) *MultiError {
	return &MultiError{
		Errors: []error{},
		For: source,
	}
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
	fmt.Fprintf(&buf, "MultiError has %d error(s) for source: %s\n", len(err.Errors), src)

	for _, err := range err.Errors {
		fmt.Fprintf(&buf, "%s\n", err.Error())
	}

	return buf.String()
}