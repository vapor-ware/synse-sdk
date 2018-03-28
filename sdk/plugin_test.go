package sdk

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vapor-ware/synse-sdk/sdk/config"
)

// File level global test configuration.
var testConfig = config.PluginConfig{
	Name:    "test config",
	Version: "test config v1",
	Network: config.NetworkSettings{
		Type:    "tcp",
		Address: "test_config",
	},
	Settings: config.Settings{
		Read:        config.ReadSettings{Buffer: 1024},
		Write:       config.WriteSettings{Buffer: 1024},
		Transaction: config.TransactionSettings{TTL: "2s"},
	},
}

// TestNewPluginNilHandlers tests creating a new Plugin with nil handlers
func TestNewPluginNilHandlers(t *testing.T) {
	_, err := NewPlugin(nil, &testConfig)
	assert.Error(t, err)
}

// TestNewPlugin tests creating a new Plugin with the required handlers defined.
func TestNewPlugin(t *testing.T) {
	// Create valid handlers for the Plugin.
	h, err := NewHandlers(testDeviceIdentifier, nil)
	assert.NoError(t, err)

	// Create the plugin.
	p, err := NewPlugin(h, &testConfig)
	assert.NoError(t, err)

	assert.Nil(t, p.server, "server should not be initialized with new plugin")
	assert.Nil(t, p.dataManager, "data manager should not be initialized with new plugin")
	assert.Nil(t, p.handlers.DeviceEnumerator)

	assert.Equal(t, &h.DeviceIdentifier, &p.handlers.DeviceIdentifier)
	assert.NotNil(t, p.Config, "plugin should be configured on init")
}

// TestNewPluginWithConfig tests creating a new Plugin and passing a valid
// config in as an argument.
func TestNewPluginWithConfig(t *testing.T) {
	// Create valid handlers for the Plugin.
	h, err := NewHandlers(testDeviceIdentifier, nil)
	assert.NoError(t, err)

	// Create a configuration for the Plugin.
	c := config.PluginConfig{
		Name:    "test-plugin",
		Version: "1.0",
		Network: config.NetworkSettings{
			Type:    "tcp",
			Address: ":666",
		},
	}

	// Create the plugin.
	p, err := NewPlugin(h, &c)
	assert.NoError(t, err)
	assert.NotNil(t, p.Config, "plugin should be configured")
}

// TestNewPluginWithIncompleteConfig tests creating a new Plugin and passing in
// an incomplete PluginConfig instance as an argument. The constructor should not
// return an error with a bad/incomplete config.
func TestNewPluginWithIncompleteConfig(t *testing.T) {
	// Create valid handlers for the Plugin.
	h, err := NewHandlers(testDeviceIdentifier, nil)
	assert.NoError(t, err)

	// network spec missing but required
	c := config.PluginConfig{
		Name:    "test-plugin",
		Version: "1.0",
	}

	// Create the plugin.
	_, err = NewPlugin(h, &c)
	assert.NoError(t, err)
}

// TestPlugin_Configure tests configuring a Plugin using a config file
// specified via environment variable.
func TestPlugin_Configure(t *testing.T) {
	cfgFilePath := "tmp/config.yml"
	cfg := `name: test-plugin
version: 1.0.0
debug: true
network:
  type: tcp
  address: ":50051"
settings:
  loop_delay: 100
  read:
    buffer_size: 150
  write:
    buffer_size: 150
    per_loop: 4
  transaction:
    ttl: 600s`

	err := writeConfigFile(cfgFilePath, cfg)
	assert.NoError(t, err)

	defer func() {
		err = os.RemoveAll("tmp")
		assert.NoError(t, err)
	}()

	os.Setenv("PLUGIN_CONFIG", "tmp")

	// Create valid handlers for the Plugin.
	h, err := NewHandlers(testDeviceIdentifier, nil)
	assert.NoError(t, err)

	// Create the plugin.
	p, err := NewPlugin(h, nil)
	assert.NoError(t, err)

	assert.NotNil(t, p.Config, "plugin is not configured, but should be")
	assert.Equal(t, "test-plugin", p.Config.Name)
	assert.Equal(t, "1.0.0", p.Config.Version)
}

// TestPlugin_setup tests setting up a Plugin successfully. This means that
// the state is validated, the devices are registered, and the Plugin components
// (server, data manager) are created.
func TestPlugin_setup(t *testing.T) {
	// Create valid handlers for the Plugin.
	h, err := NewHandlers(testDeviceIdentifier, testDeviceEnumerator)
	assert.NoError(t, err)

	// Create the plugin.
	p, err := NewPlugin(h, &testConfig)
	assert.NoError(t, err)

	// CONSIDER: Can we move setup functionality to the constructor?
	err = p.setup()
	assert.NoError(t, err)

	assert.NotNil(t, p.server, "server should be initialized on setup")
	assert.NotNil(t, p.dataManager, "data manager should be initialized on setup")
}

// TestPlugin_setup2 tests setting up a Plugin unsuccessfully. This means that
// the state is validated, the devices are registered, and the Plugin components
// (server, data manager) are created. In this case, handler validation should fail.
func TestPlugin_setup2(t *testing.T) {
	// Create invalid handlers for the plugin.
	h := Handlers{}
	p, err := NewPlugin(&h, &testConfig)
	assert.NoError(t, err)

	err = p.setup()
	assert.Error(t, err)
}
