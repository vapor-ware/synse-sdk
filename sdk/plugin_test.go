package sdk

import (
	"os"
	"testing"

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

func TestNewPluginNilHandlers(t *testing.T) {
	_, err := NewPlugin(nil, &testConfig)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestNewPlugin(t *testing.T) {
	// Create valid handlers for the Plugin.
	h, err := NewHandlers(testDeviceIdentifier, nil)
	if err != nil {
		t.Error(err)
	}

	// Create the plugin.
	p, err := NewPlugin(h, &testConfig)
	if err != nil {
		t.Error(err)
	}

	if p.server != nil {
		t.Error("plugin server should not be initialized with new plugin")
	}
	if p.handlers.DeviceEnumerator != nil {
		t.Error("device enumerator handler did not match expected")
	}
	if &p.handlers.DeviceIdentifier != &h.DeviceIdentifier {
		t.Error("device identifier handler did not match expected")
	}
	if p.dataManager != nil {
		t.Error("plugin data manager should not be initialized with new plugin")
	}
	if p.Config == nil {
		t.Error("plugin should be configured on initialization")
	}
}

func TestPlugin_SetConfig(t *testing.T) {
	// Create valid handlers for the Plugin.
	h, err := NewHandlers(testDeviceIdentifier, nil)
	if err != nil {
		t.Error(err)
	}

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
	if err != nil {
		t.Error(err)
	}

	if p.Config == nil {
		t.Error("plugin should be configured")
	}
}

func TestPlugin_SetConfig2(t *testing.T) {
	// test passing a bad configuration
	// Create valid handlers for the Plugin.
	h, err := NewHandlers(testDeviceIdentifier, nil)
	if err != nil {
		t.Error(err)
	}

	// socket spec missing but required
	c := config.PluginConfig{
		Name:    "test-plugin",
		Version: "1.0",
	}

	// Create the plugin.
	_, err = NewPlugin(h, &c)
	if err != nil {
		t.Error("expected error when setting config, but got none")
	}
}

func TestPlugin_Configure(t *testing.T) {
	// test configuring using ENV
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
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll("tmp")
		if err != nil {
			t.Error(err)
		}
	}()

	os.Setenv("PLUGIN_CONFIG", "tmp")
	// Create valid handlers for the Plugin.
	h, err := NewHandlers(testDeviceIdentifier, nil)
	if err != nil {
		t.Error(err)
	}

	// Create the plugin.
	p, err := NewPlugin(h, nil)
	if err != nil {
		t.Error(err)
	}

	if p.Config == nil {
		t.Error("plugin is not set as configured, but should be")
	}

	if p.Config.Name != "test-plugin" {
		t.Error("plugin config was not properly set")
	}
}

func TestPlugin_setup(t *testing.T) {
	// setup and validation is good
	// Create valid handlers for the Plugin.
	h, err := NewHandlers(testDeviceIdentifier, testDeviceEnumerator)
	if err != nil {
		t.Error(err)
	}

	// Create the plugin.
	p, err := NewPlugin(h, &testConfig)
	if err != nil {
		t.Error(err)
	}

	// CONSIDER: Can we move setup functionality to the constructor?
	err = p.setup()
	if err != nil {
		t.Error(err)
	}

	if p.server == nil {
		t.Error("upon setup, plugin server should be initialized")
	}
	if p.dataManager == nil {
		t.Error("upon setup, plugin device manager should be initialized")
	}
}

func TestPlugin_setup2(t *testing.T) {
	// validate handlers gives error
	h := Handlers{}
	p, err := NewPlugin(&h, &testConfig)
	if err != nil {
		t.Error(err)
	}
	p.Config = &config.PluginConfig{}

	err = p.setup()
	if err == nil {
		t.Error("expected error due to bad handlers, but got no error")
	}
}

func TestPlugin_setup3(t *testing.T) {
	// Was plugin not yet configured, but now it is configured.
	// Create valid handlers for the Plugin.
	h, err := NewHandlers(testDeviceIdentifier, testDeviceEnumerator)
	if err != nil {
		t.Error(err)
	}

	// Create the plugin.
	p, err := NewPlugin(h, &testConfig)
	if err != nil {
		t.Error(err)
	}

	err = p.setup()
	if err != nil {
		t.Error("Expected no error. Plugin configured by NewPlugin.")
	}
}
