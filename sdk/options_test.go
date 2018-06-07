package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/sdk/config"
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
		func(data map[string]interface{}) ([]*config.DeviceConfig, error) {
			return []*config.DeviceConfig{}, nil
		},
	)
	ctx := PluginContext{}
	assert.Nil(t, ctx.dynamicDeviceConfigRegistrar)

	opt(&ctx)
	assert.NotNil(t, ctx.dynamicDeviceConfigRegistrar)
}
