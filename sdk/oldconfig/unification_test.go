package oldconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUnifyDeviceConfigs_NoConfigs tests unifying configs when no
// configs are given.
func TestUnifyDeviceConfigs_NoConfigs(t *testing.T) {
	ctx, err := UnifyDeviceConfigs([]*Context{})
	assert.Error(t, err)
	assert.Nil(t, ctx)
}

// TestUnifyDeviceConfigs_NoDeviceConfig tests unifying configs when the
// contexts specified do not contain DeviceConfigs.
func TestUnifyDeviceConfigs_NoDeviceConfig(t *testing.T) {
	ctx, err := UnifyDeviceConfigs([]*Context{
		{
			Source: "test",
			Config: &PluginConfig{},
		},
	})

	assert.Error(t, err)
	assert.Nil(t, ctx)
}

// TestUnifyDeviceConfigs tests unifying configs when there is only one config
// to unify.
func TestUnifyDeviceConfigs(t *testing.T) {
	ctx, err := UnifyDeviceConfigs([]*Context{
		{
			Source: "test",
			Config: &DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations: []*Location{
					{
						Name:  "test",
						Rack:  &LocationData{Name: "rack"},
						Board: &LocationData{Name: "board"},
					},
				},
				Devices: []*DeviceKind{
					{Name: "test-device"},
				},
			},
		},
	})

	assert.NoError(t, err)
	assert.True(t, ctx.IsDeviceConfig())
	cfg := ctx.Config.(*DeviceConfig)
	assert.Equal(t, 1, len(cfg.Devices))
	assert.Equal(t, 1, len(cfg.Locations))
}

// TestUnifyDeviceConfigs2 tests unifying configs when there are multiple
// configs to unify.
func TestUnifyDeviceConfigs2(t *testing.T) {
	ctx, err := UnifyDeviceConfigs([]*Context{
		{
			Source: "test",
			Config: &DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations: []*Location{
					{
						Name:  "loc-1",
						Rack:  &LocationData{Name: "rack"},
						Board: &LocationData{Name: "board"},
					},
				},
				Devices: []*DeviceKind{
					{Name: "test-device-1"},
				},
			},
		},
		{
			Source: "test",
			Config: &DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations: []*Location{
					{
						Name:  "loc-2",
						Rack:  &LocationData{Name: "rack"},
						Board: &LocationData{Name: "board"},
					},
					{
						Name:  "loc-3",
						Rack:  &LocationData{Name: "rack"},
						Board: &LocationData{Name: "board"},
					},
				},
				Devices: []*DeviceKind{
					{Name: "test-device-2"},
				},
			},
		},
		{
			Source: "test",
			Config: &DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations: []*Location{
					{
						Name:  "loc-4",
						Rack:  &LocationData{Name: "rack"},
						Board: &LocationData{Name: "board"},
					},
				},
				Devices: []*DeviceKind{
					{Name: "test-device-3"},
					{Name: "test-device-4"},
				},
			},
		},
	})

	assert.NoError(t, err)
	assert.True(t, ctx.IsDeviceConfig())
	cfg := ctx.Config.(*DeviceConfig)
	assert.Equal(t, 4, len(cfg.Devices))
	assert.Equal(t, 4, len(cfg.Locations))
}
