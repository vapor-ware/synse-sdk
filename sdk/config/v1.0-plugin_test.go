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
  loop_delay: 100
  read:
    buffer_size: 150
  write:
    buffer_size: 150
    per_loop: 4
  transaction:
    ttl: 600`)

	v := viper.New()
	v.SetConfigType("yaml")
	v.ReadConfig(bytes.NewBuffer(config))

	fmt.Printf("%#v\n", v.AllSettings())

	handler := v1PluginConfigHandler{}

	cfg, err := handler.processPluginConfig(v)
	if err != nil {
		t.Error(err)
	}

	if cfg.Debug != true {
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
	if cfg.Settings.LoopDelay != 100 {
		t.Errorf("expected config 'settings.loop_delay' to be 100, but was %v", cfg.Settings.LoopDelay)
	}
	if cfg.Settings.Read.BufferSize != 150 {
		t.Errorf("expected config 'settings.read.buffer_size' to be 150, but was %v", cfg.Settings.Read.BufferSize)
	}
	if cfg.Settings.Write.BufferSize != 150 {
		t.Errorf("expected config 'settings.write.buffer_size' to be 150, but was %v", cfg.Settings.Write.BufferSize)
	}
	if cfg.Settings.Write.PerLoop != 4 {
		t.Errorf("expected config 'settings.write.per_loop' to be 4, but was %v", cfg.Settings.Write.PerLoop)
	}
	if cfg.Settings.Transaction.TTL != 600 {
		t.Errorf("expected config 'settings.transaction.ttl' to be 600, but was %v", cfg.Settings.Transaction.TTL)
	}
}
