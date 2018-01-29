package sdk

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/vapor-ware/synse-sdk/sdk/config"
)

func writeConfigFile(path string) error {
	_, err := os.Stat(filepath.Dir(path))
	if os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(path), os.ModePerm)
	}

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
    ttl: 600`

	return ioutil.WriteFile(path, []byte(cfg), 0644)
}

func TestNewPlugin(t *testing.T) {
	p := NewPlugin()

	if p.server != nil {
		t.Error("plugin server should not be initialized with new plugin")
	}
	if p.handlers.DeviceEnumerator != nil {
		t.Error("device enumerator handler did not match expected")
	}
	if p.handlers.DeviceIdentifier != nil {
		t.Error("device identifier handler did not match expected")
	}
	if p.dm != nil {
		t.Error("plugin data manager should not be initialized with new plugin")
	}
	if p.Config != nil {
		t.Error("plugin should not be configured on initialization")
	}
}

func TestPlugin_SetConfig(t *testing.T) {
	p := NewPlugin()

	c := config.PluginConfig{
		Name:    "test-plugin",
		Version: "1.0",
		Network: config.NetworkSettings{
			Type:    "tcp",
			Address: ":666",
		},
	}

	err := p.SetConfig(&c)
	if err != nil {
		t.Error(err)
	}

	if p.Config == nil {
		t.Error("plugin should be configured")
	}
}

func TestPlugin_SetConfig2(t *testing.T) {
	// test passing a bad configuration
	p := NewPlugin()

	// socket spec missing but required
	c := config.PluginConfig{
		Name:    "test-plugin",
		Version: "1.0",
	}

	err := p.SetConfig(&c)
	if err == nil {
		t.Error("expected error when setting config, but got none")
	}
}

func TestPlugin_Configure(t *testing.T) {
	// test configuring using ENV
	cfgFile := "tmp/config.yml"
	err := writeConfigFile(cfgFile)
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
	p := Plugin{}

	err = p.Configure()
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
	h := Handlers{
		DeviceIdentifier: testDeviceIdentifier,
		DeviceEnumerator: testDeviceEnumerator,
	}
	p := NewPlugin()
	p.RegisterHandlers(&h)
	p.Config = &config.PluginConfig{}

	err := p.setup()
	if err != nil {
		t.Error(err)
	}

	if p.server == nil {
		t.Error("upon setup, plugin server should be initialized")
	}
	if p.dm == nil {
		t.Error("upon setup, plugin device manager should be initialized")
	}
}

func TestPlugin_setup2(t *testing.T) {
	// validate handlers gives error
	h := Handlers{}
	p := NewPlugin()
	p.RegisterHandlers(&h)
	p.Config = &config.PluginConfig{}

	err := p.setup()
	if err == nil {
		t.Error("expected error due to bad handlers, but got no error")
	}
}

func TestPlugin_setup3(t *testing.T) {
	// plugin not yet configured
	h := Handlers{
		DeviceIdentifier: testDeviceIdentifier,
		DeviceEnumerator: testDeviceEnumerator,
	}
	p := NewPlugin()
	p.RegisterHandlers(&h)

	err := p.setup()
	if err == nil {
		t.Error("expected error due to plugin not being configured, but got no error")
	}
}
