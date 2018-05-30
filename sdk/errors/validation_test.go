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

func TestNewFieldNotSupportedError(t *testing.T) {
	err := NewFieldNotSupportedError("test", "foo", "3.0", "2.0")

	assert.IsType(t, &FieldNotSupported{}, err)
	assert.Equal(t, "test", err.source)
	assert.Equal(t, "foo", err.field)
	assert.Equal(t, "3.0", err.fieldVersion)
	assert.Equal(t, "2.0", err.configVersion)
}

func TestFieldNotSupported_Error(t *testing.T) {
	err := NewFieldNotSupportedError("test", "foo", "3.0", "2.0")
	out := err.Error()

	assert.Equal(t, "validating config test: field 'foo' not supported in v2.0 (field added in v3.0)", out)
}

func TestNewFieldRemovedError(t *testing.T) {
	err := NewFieldRemovedError("test", "foo", "2.0", "1.0")

	assert.IsType(t, &FieldRemoved{}, err)
	assert.Equal(t, "test", err.source)
	assert.Equal(t, "foo", err.field)
	assert.Equal(t, "2.0", err.fieldVersion)
	assert.Equal(t, "1.0", err.configVersion)
}

func TestFieldRemoved_Error(t *testing.T) {
	err := NewFieldRemovedError("test", "foo", "2.0", "3.0")
	out := err.Error()

	assert.Equal(t, "validating config test: field 'foo' not supported in v3.0 (field removed in v2.0)", out)
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
