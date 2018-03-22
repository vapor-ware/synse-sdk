package config

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spf13/viper"
)

func TestV1PluginConfigHandlerProcessPluginConfig(t *testing.T) {
	config := []byte(`version: 1
name: example
debug: true
network:
  type: unix
  address: example.sock
settings:
  mode: serial
  read:
    buffer: 150
  write:
    buffer: 150
    max: 4
  transaction:
    ttl: 600s`)

	v := viper.New()
	v.SetConfigType("yaml")
	v.ReadConfig(bytes.NewBuffer(config))

	fmt.Printf("%#v\n", v.AllSettings())

	handler := v1PluginConfigHandler{}

	cfg, err := handler.processPluginConfig(v)
	if err != nil {
		t.Error(err)
	}

	if !cfg.Debug {
		t.Errorf("expected config 'debug' to be 'true', but was %v", cfg.Debug)
	}
	if cfg.Name != "example" {
		t.Errorf("expected config 'name' to be 'example', but was %v", cfg.Name)
	}
	if cfg.Network.Type != "unix" {
		t.Errorf("expected config 'network.type' to be 'unix', but was %v", cfg.Network.Type)
	}
	if cfg.Network.Address != "example.sock" {
		t.Errorf("expected config 'network.address' to be 'example.sock', but was %v", cfg.Network.Address)
	}
	if cfg.Settings.Mode != "serial" {
		t.Errorf("expected config 'settings.mode' to be 'serial', but was %v", cfg.Settings.Mode)
	}
	if cfg.Settings.Read.Buffer != 150 {
		t.Errorf("expected config 'settings.read.buffer_size' to be 150, but was %v", cfg.Settings.Read.Buffer)
	}
	if cfg.Settings.Write.Buffer != 150 {
		t.Errorf("expected config 'settings.write.buffer_size' to be 150, but was %v", cfg.Settings.Write.Buffer)
	}
	if cfg.Settings.Write.Max != 4 {
		t.Errorf("expected config 'settings.write.per_loop' to be 4, but was %v", cfg.Settings.Write.Max)
	}
	if cfg.Settings.Transaction.TTL != "600s" {
		t.Errorf("expected config 'settings.transaction.ttl' to be 600, but was %v", cfg.Settings.Transaction.TTL)
	}
}
