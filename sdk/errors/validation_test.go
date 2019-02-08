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
