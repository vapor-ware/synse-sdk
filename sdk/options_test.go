package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCustomDeviceIdentifier tests creating a PluginOption for a custom
// device identifier.
func TestCustomDeviceIdentifier(t *testing.T) {
	opt := CustomDeviceIdentifier(
		func(data map[string]interface{}) string {
			return "foo"
		},
	)
	ctx := PluginContext{}
	assert.Nil(t, ctx.deviceIdentifier)

	opt(&ctx)
	assert.NotNil(t, ctx.deviceIdentifier)
}

// TestCustomDynamicDeviceRegistration tests creating a PluginOption for
// a custom device registration function.
func TestCustomDynamicDeviceRegistration(t *testing.T) {
	opt := CustomDynamicDeviceRegistration(
		func(data map[string]interface{}) ([]*Device, error) {
			return []*Device{}, nil
		},
	)
	ctx := PluginContext{}
	assert.Nil(t, ctx.dynamicDeviceRegistrar)

	opt(&ctx)
	assert.NotNil(t, ctx.dynamicDeviceRegistrar)
}

// TestCustomDynamicDeviceConfigRegistration tests creating a PluginOption
// for a custom device config registration function.
func TestCustomDynamicDeviceConfigRegistration(t *testing.T) {
	opt := CustomDynamicDeviceConfigRegistration(
		func(data map[string]interface{}) ([]*DeviceConfig, error) {
			return []*DeviceConfig{}, nil
		},
	)
	ctx := PluginContext{}
	assert.Nil(t, ctx.dynamicDeviceConfigRegistrar)

	opt(&ctx)
	assert.NotNil(t, ctx.dynamicDeviceConfigRegistrar)
}

// TestCustomDeviceDataValidator tests creating a PluginOption for a custom
// device data validator function.
func TestCustomDeviceDataValidator(t *testing.T) {
	opt := CustomDeviceDataValidator(
		func(i map[string]interface{}) error {
			return nil
		},
	)
	ctx := PluginContext{}
	assert.Nil(t, ctx.deviceDataValidator)

	opt(&ctx)
	assert.NotNil(t, ctx.deviceDataValidator)
}
