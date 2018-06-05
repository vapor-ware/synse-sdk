package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewConfigContext tests creating a new ConfigContext.
func TestNewConfigContext(t *testing.T) {
	var testTable = []struct {
		desc   string
		source string
		config ConfigBase
	}{
		{
			desc:   "Config is a pointer to a DeviceConfig",
			source: "test",
			config: &DeviceConfig{},
		},
		{
			desc:   "Config is a pointer to a PluginConfig",
			source: "test",
			config: &PluginConfig{},
		},
	}

	for _, testCase := range testTable {
		ctx := NewConfigContext(testCase.source, testCase.config)
		assert.NotNil(t, ctx, testCase.desc)
		assert.IsType(t, &ConfigContext{}, ctx, testCase.desc)
		assert.Equal(t, testCase.source, ctx.Source, testCase.desc)
		assert.Equal(t, testCase.config, ctx.Config, testCase.desc)
	}
}

// TestConfigContext_IsDeviceConfig tests whether the Config member is a DeviceConfig.
func TestConfigContext_IsDeviceConfig(t *testing.T) {
	var testTable = []struct {
		desc     string
		isDevCfg bool
		config   ConfigBase
	}{
		{
			desc:     "Config is a pointer to a DeviceConfig",
			isDevCfg: true,
			config:   &DeviceConfig{},
		},
		{
			desc:     "Config is a pointer to a PluginConfig",
			isDevCfg: false,
			config:   &PluginConfig{},
		},
		{
			desc:     "Config is a pointer to an OutputType",
			isDevCfg: false,
			config:   &OutputType{},
		},
	}

	for _, testCase := range testTable {
		ctx := ConfigContext{
			Source: "test",
			Config: testCase.config,
		}

		actual := ctx.IsDeviceConfig()
		assert.Equal(t, testCase.isDevCfg, actual, testCase.desc)
	}
}

// TestConfigContext_IsPluginConfig tests whether the Config member is a PluginConfig.
func TestConfigContext_IsPluginConfig(t *testing.T) {
	var testTable = []struct {
		desc        string
		isPluginCfg bool
		config      ConfigBase
	}{
		{
			desc:        "Config is a pointer to a DeviceConfig",
			isPluginCfg: false,
			config:      &DeviceConfig{},
		},
		{
			desc:        "Config is a pointer to a PluginConfig",
			isPluginCfg: true,
			config:      &PluginConfig{},
		},
		{
			desc:        "Config is a pointer to an OutputType",
			isPluginCfg: false,
			config:      &OutputType{},
		},
	}

	for _, testCase := range testTable {
		ctx := ConfigContext{
			Source: "test",
			Config: testCase.config,
		}

		actual := ctx.IsPluginConfig()
		assert.Equal(t, testCase.isPluginCfg, actual, testCase.desc)
	}
}

// TestConfigContext_IsOutputTypeConfig tests whether the Config member is an OutputType config.
func TestConfigContext_IsOutputTypeConfig(t *testing.T) {
	var testTable = []struct {
		desc         string
		isOutputType bool
		config       ConfigBase
	}{
		{
			desc:         "Config is a pointer to an OutputType",
			isOutputType: true,
			config:       &OutputType{},
		},
		{
			desc:         "Config is a pointer to a PluginConfig",
			isOutputType: false,
			config:       &PluginConfig{},
		},
		{
			desc:         "Config is a pointer to a DeviceConfig",
			isOutputType: false,
			config:       &DeviceConfig{},
		},
	}

	for _, testCase := range testTable {
		ctx := ConfigContext{
			Source: "test",
			Config: testCase.config,
		}

		actual := ctx.IsOutputTypeConfig()
		assert.Equal(t, testCase.isOutputType, actual, testCase.desc)
	}
}
