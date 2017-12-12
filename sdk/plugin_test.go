package sdk

import (
	"testing"
)

func TestNewPlugin(t *testing.T) {
	h := Handlers{}
	p := NewPlugin(&h)

	if p.server != nil {
		t.Error("plugin server should not be initialized with new plugin")
	}
	if p.handlers != &h {
		t.Error("handlers did not match expected")
	}
	if p.dm != nil {
		t.Error("plugin data manager should not be initialized with new plugin")
	}
	if p.isConfigured {
		t.Error("plugin should not be configured on initialization")
	}
}

func TestPlugin_SetConfig(t *testing.T) {
	h := Handlers{}
	p := NewPlugin(&h)

	c := PluginConfig{
		Name:    "test-plugin",
		Version: "1.0",
		Socket: PluginConfigSocket{
			Network: "tcp",
			Address: ":666",
		},
	}

	err := p.SetConfig(&c)
	if err != nil {
		t.Error(err)
	}

	if !p.isConfigured {
		t.Error("plugin should be configured")
	}
}

func TestPlugin_SetConfig2(t *testing.T) {
	// test passing a bad configuration
	h := Handlers{}
	p := NewPlugin(&h)

	// socket spec missing but required
	c := PluginConfig{
		Name:    "test-plugin",
		Version: "1.0",
	}

	err := p.SetConfig(&c)
	if err == nil {
		t.Error("expected error when setting config, but got none")
	}
}

func TestPlugin_Configure(t *testing.T) {
	// test configuring with the default config location
}

func TestPlugin_Configure2(t *testing.T) {
	// test configuring using ENV
}

func TestPlugin_setup(t *testing.T) {
	// setup and validation is good

}

func TestPlugin_setup2(t *testing.T) {
	// validate handlers gives error

}

func TestPlugin_setup3(t *testing.T) {
	// plugin not yet configured

}
