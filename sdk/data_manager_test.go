package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vapor-ware/synse-sdk/sdk/config"
)

// DeviceID gets the unique identifiers out of the plugin-specific
// configuration to be used in UID generation. Required to construct Handlers.
func DeviceID(data map[string]string) string {
	return data["id"]
}

// TestNewDataManager tests creating a new dataManager instance successfully.
func TestNewDataManager(t *testing.T) {
	// Create handlers.
	h, err := NewHandlers(DeviceID, nil)
	assert.NoError(t, err)

	c := config.PluginConfig{
		Name:    "test",
		Version: "test",
		Network: config.NetworkSettings{
			Type:    "tcp",
			Address: "test",
		},
		Settings: config.Settings{
			Read:        config.ReadSettings{Buffer: 200},
			Write:       config.WriteSettings{Buffer: 200},
			Transaction: config.TransactionSettings{TTL: "2s"},
		},
	}
	p := Plugin{handlers: h}
	err = p.SetConfig(&c)
	assert.NoError(t, err)

	d, err := newDataManager(&p)
	assert.NoError(t, err)

	assert.Equal(t, 200, cap(d.writeChannel))
	assert.Equal(t, 200, cap(d.readChannel))
	assert.Equal(t, h, d.handlers)
}

// TestNewDataManager2 tests creating a new dataManager instance successfully with
// a different configuration.
func TestNewDataManager2(t *testing.T) {
	// Create handlers.
	h, err := NewHandlers(DeviceID, nil)
	assert.NoError(t, err)

	c := &config.PluginConfig{
		Name:    "test",
		Version: "test",
		Network: config.NetworkSettings{
			Type:    "tcp",
			Address: "test",
		},
		Settings: config.Settings{
			Read:        config.ReadSettings{Buffer: 500},
			Write:       config.WriteSettings{Buffer: 500},
			Transaction: config.TransactionSettings{TTL: "2s"},
		},
	}
	p := Plugin{handlers: h}
	err = p.SetConfig(c)
	assert.NoError(t, err)

	// Create the dataManager
	d, err := newDataManager(&p)
	assert.NoError(t, err)

	assert.Equal(t, 500, cap(d.writeChannel))
	assert.Equal(t, 500, cap(d.readChannel))
	assert.Equal(t, h, d.handlers)
}
