package sdk

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInvalidArgumentErr tests constructing a new InvalidArgument error.
func TestInvalidArgumentErr(t *testing.T) {
	errString := "test error"
	err := invalidArgumentErr(errString)

	assert.True(t, strings.Contains(err.Error(), "InvalidArgument"))
	assert.True(t, strings.Contains(err.Error(), errString))
}

// TestNotFoundErr tests constructing a new NotFound error.
func TestNotFoundErr(t *testing.T) {
	errString := "test error"
	err := notFoundErr(errString)

	assert.True(t, strings.Contains(err.Error(), "NotFound"))
	assert.True(t, strings.Contains(err.Error(), errString))
}
