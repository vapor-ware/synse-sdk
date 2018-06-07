package config

import (
	"time"

	"github.com/creasty/defaults"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
)

const (
	modeSerial   = "serial"
	modeParallel = "parallel"

	networkTypeTCP  = "tcp"
	networkTypeUnix = "unix"
)

var (
	// The current (latest) version of the plugin config scheme.
	currentPluginSchemeVersion = "1.0"
)

// NewDefaultPluginConfig creates a new instance of a PluginConfig with its
// default values resolved.
func NewDefaultPluginConfig() (*PluginConfig, error) {
	config := &PluginConfig{
		SchemeVersion: SchemeVersion{Version: currentPluginSchemeVersion},
	}
	err := defaults.Set(config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// PluginConfig contains the configuration options for the plugin.
type PluginConfig struct {

	// SchemeVersion is the version of the configuration scheme.
	SchemeVersion `yaml:",inline"`

	// Debug is a flag that determines whether the plugin should run
	// with debug logging or not.
	Debug bool `default:"false" yaml:"debug,omitempty" addedIn:"1.0"`

	// Settings provide specifications for how the plugin should run.
	Settings *PluginSettings `default:"{}" yaml:"settings,omitempty" addedIn:"1.0"`

	// Network specifies the networking configuration for the plugin.
	Network *NetworkSettings `default:"{}" yaml:"network,omitempty" addedIn:"1.0"`

	// DynamicRegistration specifies configuration settings and data
	// for how the plugin should handle dynamic device registration.
	DynamicRegistration *DynamicRegistrationSettings `default:"{}" yaml:"dynamicRegistration,omitempty" addedIn:"1.0"`

	// Limiter specifies settings for a rate limiter for reads/writes.
	Limiter *LimiterSettings `yaml:"limiter,omitempty" addedIn:"1.0"`

	// Health sepcifies the settings for health checking in the plugin.
	Health *HealthSettings `default:"{}" yaml:"health,omitempty" addedIn:"1.0"`

	// Context is a map that allows the plugin to specify any arbitrary
	// data it may need.
	Context map[string]interface{} `default:"{}" yaml:"context,omitempty" addedIn:"1.0"`
}

// Validate validates that the PluginConfig has no configuration errors.
func (config PluginConfig) Validate(multiErr *errors.MultiError) {
	// A version must be specified and it must be of the correct format.
	_, err := config.GetVersion()
	if err != nil {
		// TODO -- using multiErr.Context["source"] assumes that all of the
		// configs came from file. Need to see if there is a way to check
		// viper for whether or not we know if the source is file or commandline.
		multiErr.Add(errors.NewValidationError(multiErr.Context["source"], err.Error()))
	}

	// If network is nil or an empty struct, error. We need to know how
	// the plugin should communicate with Synse Server.
	if config.Network == nil || config.Network == (&NetworkSettings{}) {
		multiErr.Add(errors.NewFieldRequiredError(multiErr.Context["source"], "network"))
	}
}

// PluginSettings specifies the configuration options that determine the
// runtime behavior of the plugin.
type PluginSettings struct {
	// Mode is the run mode of the read and write loops. This can either
	// be "serial" or "parallel".
	Mode string `default:"serial" yaml:"mode,omitempty" addedIn:"1.0"`

	// Read contains the settings to configure read behavior.
	Read *ReadSettings `default:"{}" yaml:"read,omitempty" addedIn:"1.0"`

	// Write contains the settings to configure write behavior.
	Write *WriteSettings `default:"{}" yaml:"write,omitempty" addedIn:"1.0"`

	// Transaction contains the settings to configure transaction
	// handling behavior.
	Transaction *TransactionSettings `default:"{}" yaml:"transaction,omitempty" addedIn:"1.0"`
}

// Validate validates that the PluginSettings has no configuration errors.
func (settings PluginSettings) Validate(multiErr *errors.MultiError) {
	if settings.Mode != modeSerial && settings.Mode != modeParallel {
		multiErr.Add(errors.NewInvalidValueError(
			multiErr.Context["source"],
			"settings.mode",
			"one of: serial, parallel",
		))
	}
}

// IsSerial checks if the PluginSettings is configured with mode "serial".
func (settings *PluginSettings) IsSerial() bool {
	return settings.Mode == modeSerial
}

// IsParallel checks if the PluginSettings is configured with mode "parallel".
func (settings *PluginSettings) IsParallel() bool {
	return settings.Mode == modeParallel
}

// NetworkSettings specifies the configuration options around the gRPC
// server's networking behavior.
type NetworkSettings struct {
	// Type is the type of networking. Currently, this must be one of
	// "tcp" (TCP/IP) or "unix" (Unix Socket)
	Type string `yaml:"type,omitempty" addedIn:"1.0"`

	// Address is the address to communicate over. For "tcp", this would
	// be the host/port (e.g. 0.0.0.0:50001). For "unix", this would be
	// the name of the unix socket (e.g. plugin.sock).
	Address string `yaml:"address,omitempty" addedIn:"1.0"`
}

// Validate validates that the NetworkSettings has no configuration errors.
func (settings NetworkSettings) Validate(multiErr *errors.MultiError) {
	if settings.Type == "" {
		multiErr.Add(errors.NewFieldRequiredError(multiErr.Context["source"], "network.type"))
	} else {
		if settings.Type != networkTypeTCP && settings.Type != networkTypeUnix {
			multiErr.Add(errors.NewInvalidValueError(
				multiErr.Context["source"],
				"network.type",
				"one of: unix, tcp",
			))
		}
	}
	if settings.Address == "" {
		multiErr.Add(errors.NewFieldRequiredError(multiErr.Context["source"], "network.address"))
	}
}

// DynamicRegistrationSettings specifies configuration and data for
// the dynamic registration of devices.
type DynamicRegistrationSettings struct {
	// The plugin configuration for dynamic registration. This map holds the
	// plugin-specific data that can be used to dynamically register new devices.
	// As an example, this could hold the information for connecting with a server,
	// or it could contain a bus address, etc.
	Config map[string]interface{} `default:"{}" yaml:"config,omitempty" addedIn:"1.0"`
}

// Validate validates that the DynamicRegistrationSettings has no configuration errors.
func (settings DynamicRegistrationSettings) Validate(multiErr *errors.MultiError) {
	// nothing to validate here.
}

// LimiterSettings specifies configurations for a rate limiter on reads
// and writes.
type LimiterSettings struct {
	// Rate is the limit, or maximum frequency of events. A rate of
	// 0 signifies 'unlimited'.
	Rate int `yaml:"rate,omitempty" addedIn:"1.0"`

	// Burst defines the bucket size for the limiter, or maximum number
	// of events that can be fulfilled at once. If this is 0, it will take
	// the same value as the rate.
	Burst int `yaml:"burst,omitempty" addedIn:"1.0"`
}

// Validate validates that the LimiterSettings has no configuration errors.
func (settings LimiterSettings) Validate(multiErr *errors.MultiError) {
	if settings.Rate < 0 {
		multiErr.Add(errors.NewInvalidValueError(
			multiErr.Context["source"],
			"limiter.rate",
			"greater than or equal to 0",
		))
	}

	if settings.Burst < 0 {
		multiErr.Add(errors.NewInvalidValueError(
			multiErr.Context["source"],
			"limiter.burst",
			"greater than or equal to 0",
		))
	}
}

// ReadSettings provides configuration options for read operations.
type ReadSettings struct {
	// Enabled globally enables or disables reading for the plugin.
	// By default, a plugin will have reading enabled.
	Enabled bool `default:"true" yaml:"enabled,omitempty" addedIn:"1.0"`

	// Interval specifies the interval at which devices should be
	// read from. This is 1s by default.
	Interval string `default:"1s" yaml:"interval,omitempty" addedIn:"1.0"`

	// Buffer defines the size of the read buffer. This will be
	// the size of the channel that passes along read responses.
	Buffer int `default:"100" yaml:"buffer,omitempty" addedIn:"1.0"`
}

// Validate validates that the ReadSettings has no configuration errors.
func (settings ReadSettings) Validate(multiErr *errors.MultiError) {
	// Try parsing the interval to validate it is a correctly specified duration string.
	_, err := settings.GetInterval()
	if err != nil {
		multiErr.Add(errors.NewValidationError(multiErr.Context["source"], err.Error()))
	}

	// If the buffer size is set to 0, return an error. Previously, this
	// was allowed, as a size of 0 could indicate "no read", but now we
	// have the 'enabled' field, so we don't need to support this.
	if settings.Buffer <= 0 {
		multiErr.Add(errors.NewInvalidValueError(
			multiErr.Context["source"],
			"settings.read.buffer",
			"a value greater than 0",
		))
	}
}

// GetInterval gets the read interval as a duration. If the config
// has been validated successfully, this should never return an error.
func (settings *ReadSettings) GetInterval() (time.Duration, error) {
	return time.ParseDuration(settings.Interval)
}

// WriteSettings provides configuration options for write operations.
type WriteSettings struct {
	// Enabled globally enables or disables writing for the plugin.
	// By default, a plugin will have writing enabled.
	Enabled bool `default:"true" yaml:"enabled,omitempty" addedIn:"1.0"`

	// Interval specifies the interval at which devices should be
	// written to. This is 1s by default.
	Interval string `default:"1s" yaml:"interval,omitempty" addedIn:"1.0"`

	// Buffer defines the size of the write buffer. This will be
	// the size of the channel that passes along write requests.
	Buffer int `default:"100" yaml:"buffer,omitempty" addedIn:"1.0"`

	// Max is the maximum number of write transactions to process
	// in a single batch. In general, this can tune performance when
	// running in serial mode.
	Max int `default:"100" yaml:"max,omitempty" addedIn:"1.0"`
}

// Validate validates that the WriteSettings has no configuration errors.
func (settings WriteSettings) Validate(multiErr *errors.MultiError) {
	// Try parsing the interval to validate it is a correctly specified duration string.
	_, err := settings.GetInterval()
	if err != nil {
		multiErr.Add(errors.NewValidationError(multiErr.Context["source"], err.Error()))
	}

	// If the buffer size is set to 0, return an error. Previously, this
	// was allowed, as a size of 0 could indicate "no write", but now we
	// have the 'enabled' field, so we don't need to support this.
	if settings.Buffer <= 0 {
		multiErr.Add(errors.NewInvalidValueError(
			multiErr.Context["source"],
			"settings.write.buffer",
			"a value greater than 0",
		))
	}

	if settings.Max <= 0 {
		multiErr.Add(errors.NewInvalidValueError(
			multiErr.Context["source"],
			"settings.write.max",
			"a value greater than 0",
		))
	}
}

// GetInterval gets the write interval as a duration. If the config
// has been validated successfully, this should never return an error.
func (settings WriteSettings) GetInterval() (time.Duration, error) {
	return time.ParseDuration(settings.Interval)
}

// TransactionSettings provides configuration options for transaction operations.
type TransactionSettings struct {
	// TTL is the time-to-live for a transaction in the transaction cache.
	TTL string `default:"5m" yaml:"ttl,omitempty" addedIn:"1.0"`
}

// Validate validates that the TransactionSettings has no configuration errors.
func (settings TransactionSettings) Validate(multiErr *errors.MultiError) {
	// Try parsing the interval to validate it is a correctly specified duration string.
	_, err := settings.GetTTL()
	if err != nil {
		multiErr.Add(errors.NewValidationError(multiErr.Context["source"], err.Error()))
	}
}

// GetTTL gets the transaction TTL as a duration. If the config has been
// validated successfully, this should never return an error.
func (settings *TransactionSettings) GetTTL() (time.Duration, error) {
	return time.ParseDuration(settings.TTL)
}

// HealthSettings provides configuration options around health checking in
// the plugin.
type HealthSettings struct {
	// UseDefaults determines whether the plugin should use the built-in health
	// checks or not.
	UseDefaults bool `default:"true" yaml:"useDefaults,omitempty" addedIn:"1.0"`
}

// Validate validates that the HealthSettings has no configuration errors.
func (settings HealthSettings) Validate(multiErr *errors.MultiError) {
	// Nothing to validate here.
}
