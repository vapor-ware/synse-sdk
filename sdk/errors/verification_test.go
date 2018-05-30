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
