package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test_newPluginContext tests creating a new instance of a plugin context.
func Test_newPluginContext(t *testing.T) {
	context := newPluginContext()

	assert.NotNil(t, context)
	assert.IsType(t, &PluginContext{}, context)

	// Can't compare function pointers, so just make sure its not nil for noe
	assert.NotNil(t, context.deviceIdentifier)
	assert.NotNil(t, context.dynamicDeviceRegistrar)
	assert.NotNil(t, context.dynamicDeviceConfigRegistrar)
	assert.NotNil(t, context.deviceDataValidator)
}
