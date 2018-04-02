package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewHandlers tests initializing a new Handlers struct successfully.
func TestNewHandlers(t *testing.T) {
	handlers, err := NewHandlers(
		testDeviceIdentifier,
		testDeviceEnumerator,
	)
	assert.NoError(t, err)
	assert.NotNil(t, handlers)
}

// TestNewHandlersErr tests initializing a new Handlers struct unsuccessfully
// by passing in nil for the required device identifier handler.
func TestNewHandlersErr(t *testing.T) {
	handlers, err := NewHandlers(
		nil,
		testDeviceEnumerator,
	)
	assert.Error(t, err)
	assert.Nil(t, handlers)
}
