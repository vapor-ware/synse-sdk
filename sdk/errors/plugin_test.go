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
