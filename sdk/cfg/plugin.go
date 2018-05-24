package cfg

import (
	"time"

	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
)

const (
	modeSerial   = "serial"
	modeParallel = "parallel"
)

// NewPluginConfig creates a new instance of PluginConfig, populated from
// the configuration read in by Viper. This will include config options from
// the command line and from file.
func NewPluginConfig() (*PluginConfig, error) {
	// First, we setup all the lookup info for the viper instance.
	viper.SetConfigName("config")

	// Set the environment variable lookup
	viper.SetEnvPrefix("plugin")
	viper.AutomaticEnv()

	// If the PLUGIN_CONFIG environment variable is set, we will only search for
	// the config in that specified path, as we should expect the user-specified
	// value to be there. Otherwise, we will look through a set of pre-defined
	// configuration locations (in order of search):
	//  - current working directory
	//  - local config directory
	//  - the default config location in /etc
	configPath := os.Getenv(EnvPluginConfig)
	if configPath != "" {
		viper.AddConfigPath(configPath)
	} else {
		for _, path := range pluginConfigSearchPaths {
			viper.AddConfigPath(path)
		}
	}

	// Set default values for the PluginConfig
	SetDefaults()

	// will be used for the ConfigContext
	//configFile := viper.ConfigFileUsed()

	// Read in the configuration
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	config := &PluginConfig{}
	err = mapstructure.Decode(viper.AllSettings(), config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// SetDefaults sets the default values for the PluginConfig via Viper.
func SetDefaults() {
	viper.SetDefault("debug", false)
	viper.SetDefault("settings.mode", modeSerial)
	viper.SetDefault("settings.read.interval", "1s")
	viper.SetDefault("settings.read.buffer", 100)
	viper.SetDefault("settings.read.enabled", true)
	viper.SetDefault("settings.write.interval", "1s")
	viper.SetDefault("settings.write.buffer", 100)
	viper.SetDefault("settings.write.max", 100)
	viper.SetDefault("settings.write.enabled", true)
	viper.SetDefault("settings.transaction.ttl", "5m")
}

// PluginConfig contains the configuration options for the plugin.
type PluginConfig struct {

	// ConfigVersion is the version of the configuration scheme.
	ConfigVersion

	// Debug is a flag that determines whether the plugin should run
	// with debug logging or not.
	Debug bool `yaml:"debug,omitempty" addedIn:"1.0"`

	// Settings provide specifications for how the plugin should run.
	Settings *PluginSettings `yaml:"settings,omitempty" addedIn:"1.0"`

	// Network specifies the networking configuration for the plugin.
	Network *NetworkSettings `yaml:"network,omitempty" addedIn:"1.0"`

	// DynamicRegistration specifies configuration settings and data
	// for how the plugin should handle dynamic device registration.
	DynamicRegistration *DynamicRegistrationSettings `yaml:"dynamicRegistration,omitempty" addedIn:"1.0"`

	// Limiter specifies settings for a rate limiter for reads/writes.
	Limiter *LimiterSettings `yaml:"limiter,omitempty" addedIn:"1.0"`

	// Context is a map that allows the plugin to specify any arbitrary
	// data it may need.
	Context map[string]interface{} `yaml:"context,omitempty" addedIn:"1.0"`
}

// Validate validates that the PluginConfig has no configuration errors.
func (config *PluginConfig) Validate(multiErr *errors.MultiError) {
	// A version must be specified and it must be of the correct format.
	_, err := config.GetSchemeVersion()
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
	Mode string `yaml:"mode,omitempty" addedIn:"1.0"`

	// Read contains the settings to configure read behavior.
	Read ReadSettings `yaml:"read,omitempty" addedIn:"1.0"`

	// Write contains the settings to configure write behavior.
	Write WriteSettings `yaml:"write,omitempty" addedIn:"1.0"`

	// Transaction contains the settings to configure transaction
	// handling behavior.
	Transaction TransactionSettings `yaml:"transaction,omitempty" addedIn:"1.0"`
}

// Validate validates that the PluginSettings has no configuration errors.
func (settings *PluginSettings) Validate(multiErr *errors.MultiError) {
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
func (settings *NetworkSettings) Validate(multiErr *errors.MultiError) {
	if settings.Type == "" {
		multiErr.Add(errors.NewFieldRequiredError(multiErr.Context["source"], "network.type"))
	}
	if settings.Address == "" {
		multiErr.Add(errors.NewFieldRequiredError(multiErr.Context["source"], "network.address"))
	}
}

// DynamicRegistrationSettings specifies configuration and data for
// the dynamic registration of devices.
type DynamicRegistrationSettings struct {
}

// Validate validates that the DynamicRegistrationSettings has no configuration errors.
func (settings *DynamicRegistrationSettings) Validate(multiErr *errors.MultiError) {
	// todo
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
func (settings *LimiterSettings) Validate(multiErr *errors.MultiError) {
	// Nothing to validate.
}

// ReadSettings provides configuration options for read operations.
type ReadSettings struct {
	// Enabled globally enables or disables reading for the plugin.
	// By default, a plugin will have reading enabled.
	Enabled bool `yaml:"enabled,omitempty" addedIn:"1.0"`

	// Interval specifies the interval at which devices should be
	// read from. This is 1s by default.
	Interval string `yaml:"interval,omitempty" addedIn:"1.0"`

	// Buffer defines the size of the read buffer. This will be
	// the size of the channel that passes along read responses.
	Buffer int `yaml:"buffer,omitempty" addedIn:"1.0"`
}

// Validate validates that the ReadSettings has no configuration errors.
func (settings *ReadSettings) Validate(multiErr *errors.MultiError) {
	// Try parsing the interval to validate it is a correctly specified
	// duration string.
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
	Enabled bool `yaml:"enabled,omitempty" addedIn:"1.0"`

	// Interval specifies the interval at which devices should be
	// written to. This is 1s by default.
	Interval string `yaml:"interval,omitempty" addedIn:"1.0"`

	// Buffer defines the size of the write buffer. This will be
	// the size of the channel that passes along write requests.
	Buffer int `yaml:"buffer,omitempty" addedIn:"1.0"`

	// Max is the maximum number of write transactions to process
	// in a single batch. In general, this can tune performance when
	// running in serial mode.
	Max int `yaml:"max,omitempty" addedIn:"1.0"`
}

// Validate validates that the WriteSettings has no configuration errors.
func (settings *WriteSettings) Validate(multiErr *errors.MultiError) {
	// Try parsing the interval to validate it is a correctly specified
	// duration string.
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
func (settings *WriteSettings) GetInterval() (time.Duration, error) {
	return time.ParseDuration(settings.Interval)
}

// TransactionSettings provides configuration options for transaction operations.
type TransactionSettings struct {
	// TTL is the time-to-live for a transaction in the transaction cache.
	TTL string `yaml:"ttl,omitempty" addedIn:"1.0"`
}

// Validate validates that the TransactionSettings has no configuration errors.
func (settings *TransactionSettings) Validate(multiErr *errors.MultiError) {
	// Try parsing the interval to validate it is a correctly specified
	// duration string.
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
