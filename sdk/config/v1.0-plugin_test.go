package config

import (
	"bytes"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

// TestV1PluginConfigHandlerProcessPluginConfig tests successfully processing
// a v1 plugin config.
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
	err := v.ReadConfig(bytes.NewBuffer(config))
	assert.NoError(t, err)

	handler := v1PluginConfigHandler{}

	cfg, err := handler.processPluginConfig(v)
	assert.NoError(t, err)

	assert.True(t, cfg.Debug)
	assert.Equal(t, "example", cfg.Name)
	assert.Equal(t, "unix", cfg.Network.Type)
	assert.Equal(t, "example.sock", cfg.Network.Address)
	assert.Equal(t, "serial", cfg.Settings.Mode)
	assert.Equal(t, 150, cfg.Settings.Read.Buffer)
	assert.Equal(t, 150, cfg.Settings.Write.Buffer)
	assert.Equal(t, 4, cfg.Settings.Write.Max)
	assert.Equal(t, "600s", cfg.Settings.Transaction.TTL)
	assert.Nil(t, cfg.Limiter)

	ttl, err := cfg.Settings.Transaction.GetTTL()
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(600)*time.Second, ttl)
}

// TestV1PluginConfigHandlerProcessPluginConfig2 tests unsuccessfully processing
// a v1 plugin config because the auto-enumerate field is incorrectly specified.
func TestV1PluginConfigHandlerProcessPluginConfig2(t *testing.T) {
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
    ttl: 600s
auto_enumerate: invalid-value`)

	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(bytes.NewBuffer(config))
	assert.NoError(t, err)

	handler := v1PluginConfigHandler{}

	cfg, err := handler.processPluginConfig(v)
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

// TestV1PluginConfigHandlerProcessPluginConfig3 tests unsuccessfully processing
// a v1 plugin config because the context field is incorrectly specified.
func TestV1PluginConfigHandlerProcessPluginConfig3(t *testing.T) {
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
    ttl: 600s
context: invalid-value`)

	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(bytes.NewBuffer(config))
	assert.NoError(t, err)

	handler := v1PluginConfigHandler{}

	cfg, err := handler.processPluginConfig(v)
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

// TestV1PluginConfigHandlerProcessPluginConfig4 tests successfully processing
// a v1 plugin config, when there is a limiter defined.
func TestV1PluginConfigHandlerProcessPluginConfig4(t *testing.T) {
	config := []byte(`version: 1
name: example
network:
  type: unix
  address: example.sock
limiter:
  rate: 10
  burst: 10`)

	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(bytes.NewBuffer(config))
	assert.NoError(t, err)

	handler := v1PluginConfigHandler{}

	cfg, err := handler.processPluginConfig(v)
	assert.NoError(t, err)

	assert.False(t, cfg.Debug)
	assert.Equal(t, "example", cfg.Name)
	assert.Equal(t, "unix", cfg.Network.Type)
	assert.Equal(t, "example.sock", cfg.Network.Address)
	assert.Equal(t, 10, cfg.Limiter.Burst())
	assert.Equal(t, rate.Limit(10), cfg.Limiter.Limit())
}

// TestV1PluginConfigHandlerProcessPluginConfig5 tests successfully processing
// a v1 plugin config, when there is a limiter defined with a 0-valued rate.
func TestV1PluginConfigHandlerProcessPluginConfig5(t *testing.T) {
	config := []byte(`version: 1
name: example
network:
  type: unix
  address: example.sock
limiter:
  rate: 0
  burst: 10`)

	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(bytes.NewBuffer(config))
	assert.NoError(t, err)

	handler := v1PluginConfigHandler{}

	cfg, err := handler.processPluginConfig(v)
	assert.NoError(t, err)

	assert.False(t, cfg.Debug)
	assert.Equal(t, "example", cfg.Name)
	assert.Equal(t, "unix", cfg.Network.Type)
	assert.Equal(t, "example.sock", cfg.Network.Address)
	assert.Equal(t, 10, cfg.Limiter.Burst())
	assert.Equal(t, rate.Inf, cfg.Limiter.Limit()) // 0 rate indicates infinite limit
}

// TestV1PluginConfigHandlerProcessPluginConfig6 tests successfully processing
// a v1 plugin config, when there is a limiter defined with a 0-valued burst.
func TestV1PluginConfigHandlerProcessPluginConfig6(t *testing.T) {
	config := []byte(`version: 1
name: example
network:
  type: unix
  address: example.sock
limiter:
  rate: 10
  burst: 0`)

	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(bytes.NewBuffer(config))
	assert.NoError(t, err)

	handler := v1PluginConfigHandler{}

	cfg, err := handler.processPluginConfig(v)
	assert.NoError(t, err)

	assert.False(t, cfg.Debug)
	assert.Equal(t, "example", cfg.Name)
	assert.Equal(t, "unix", cfg.Network.Type)
	assert.Equal(t, "example.sock", cfg.Network.Address)
	assert.Equal(t, 10, cfg.Limiter.Burst()) // this should take the value of the limit
	assert.Equal(t, rate.Limit(10), cfg.Limiter.Limit())
}

// TestV1PluginConfigHandlerProcessPluginConfig7 tests unsuccessfully processing
// a v1 plugin config because of validation error on required fields (missing "name" field).
func TestV1PluginConfigHandlerProcessPluginConfig7(t *testing.T) {
	config := []byte(`version: 1
network:
  type: unix
  address: example.sock`)

	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(bytes.NewBuffer(config))
	assert.NoError(t, err)

	handler := v1PluginConfigHandler{}

	cfg, err := handler.processPluginConfig(v)
	assert.Error(t, err)
	assert.Nil(t, cfg)
}
