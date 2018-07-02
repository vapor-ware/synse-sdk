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

// Test_checkDeviceHandlers tests checking device handlers when there are
// no device handlers.
func Test_checkDeviceHandlers(t *testing.T) {
	context := newPluginContext()

	err := context.checkDeviceHandlers()
	assert.NoError(t, err)
}

// Test_checkDeviceHandlers2 tests checking device handlers when there are
// no duplicate device handlers.
func Test_checkDeviceHandlers2(t *testing.T) {
	context := newPluginContext()
	context.deviceHandlers = []*DeviceHandler{
		{Name: "foo"},
		{Name: "bar"},
	}

	err := context.checkDeviceHandlers()
	assert.NoError(t, err)
}

// Test_checkDeviceHandlers3 tests checking device handlers when there is
// one duplicate device handler pair.
func Test_checkDeviceHandlers3(t *testing.T) {
	context := newPluginContext()
	context.deviceHandlers = []*DeviceHandler{
		{Name: "foo"},
		{Name: "bar"},
		{Name: "foo"},
	}

	err := context.checkDeviceHandlers()
	assert.Error(t, err)
}

// Test_checkDeviceHandlers4 tests checking device handlers when there are
// multiple duplicate device handlers.
func Test_checkDeviceHandlers4(t *testing.T) {
	context := newPluginContext()
	context.deviceHandlers = []*DeviceHandler{
		{Name: "foo"},
		{Name: "bar"},
		{Name: "baz"},
		{Name: "foo"},
		{Name: "baz"},
	}

	err := context.checkDeviceHandlers()
	assert.Error(t, err)
}
