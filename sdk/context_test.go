package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test_newPluginContext tests creating a new instance of a plugin context.
func Test_newPluginContext(t *testing.T) {
	ctx := newPluginContext()

	assert.NotNil(t, ctx)
	assert.IsType(t, &PluginContext{}, ctx)

	// Can't compare function pointers, so just make sure its not nil for noe
	assert.NotNil(t, ctx.deviceIdentifier)
	assert.NotNil(t, ctx.dynamicDeviceRegistrar)
	assert.NotNil(t, ctx.dynamicDeviceConfigRegistrar)
	assert.NotNil(t, ctx.deviceDataValidator)
}
