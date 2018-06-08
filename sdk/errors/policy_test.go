package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPolicyViolationError(t *testing.T) {
	err := NewPolicyViolationError("DeviceConfigOptional", "message")

	assert.IsType(t, &PolicyViolationError{}, err)
	assert.Equal(t, "DeviceConfigOptional", err.policy)
	assert.Equal(t, "message", err.msg)
}

func TestPolicyViolationError_Error(t *testing.T) {
	err := NewPolicyViolationError("PluginConfigRequired", "message")
	out := err.Error()

	assert.Equal(t, "policy violation (PluginConfigRequired): message", out)
}
