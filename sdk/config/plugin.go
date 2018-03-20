package config

import (
	"fmt"
	"os"

	"time"

	"github.com/spf13/viper"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
	"golang.org/x/time/rate"
)

const (
	defaultConfigPath = "/etc/synse/plugin"
)

var homeConfigPath = os.Getenv("HOME") + "/.synse/plugin"

// PluginConfig specifies the configuration options for the plugin.
type PluginConfig struct {
	Name          string
	Version       string
	Debug         bool
	Settings      Settings
	Network       NetworkSettings
	AutoEnumerate []map[string]interface{}
	Context       map[string]interface{}
	Limiter       *rate.Limiter
}

// NetworkSettings specifies the configuration options surrounding the
// gRPC server's networking behavior.
type NetworkSettings struct {
	Type    string
	Address string
}

// Settings specifies the configuration options that determine the
// behavior of the plugin.
type Settings struct {
	Mode        string
	Read        ReadSettings
	Write       WriteSettings
	Transaction TransactionSettings
}

// IsSerial checks if Settings is configured with mode "serial".
func (s *Settings) IsSerial() bool {
	return s.Mode == "serial"
}

// IsParallel checks if Settings is configured with mode "parallel".
func (s *Settings) IsParallel() bool {
	return s.Mode == "parallel"
}

// ReadSettings provides configuration options for read operations.
type ReadSettings struct {
	Enabled  bool
	Interval string
	Buffer   int
}

// GetInterval gets the ReadSettings interval as a Duration.
func (setting *ReadSettings) GetInterval() (time.Duration, error) {
	return time.ParseDuration(setting.Interval)
}

// WriteSettings provides configuration options for write operations.
type WriteSettings struct {
	Enabled  bool
	Interval string
	Buffer   int
	Max      int
}

// GetInterval gets the WriteSettings interval as a Duration.
func (setting *WriteSettings) GetInterval() (time.Duration, error) {
	return time.ParseDuration(setting.Interval)
}

// TransactionSettings provides configuration options for transaction operations.
type TransactionSettings struct {
	TTL string
}

// GetTTL gets the TransactionSettings TTL as a Duration.
func (setting *TransactionSettings) GetTTL() (time.Duration, error) {
	return time.ParseDuration(setting.TTL)
}

// Validate checks the PluginConfig instance to make sure it has all of
// the required fields populated.
func (c *PluginConfig) Validate() error {
	// Config errors
	if c.Name == "" {
		return fmt.Errorf("config validation failed: missing required field 'name'")
	}

	if c.Version == "" {
		return fmt.Errorf("config validation failed: missing required field 'version'")
	}

	if c.Network.Type == "" {
		return fmt.Errorf("config validation failed: missing required field 'network.type'")
	}

	if c.Network.Address == "" {
		return fmt.Errorf("config validation failed: missing required field 'network.address'")
	}

	// Config warnings
	if c.Settings.Write.Buffer == 0 {
		logger.Warn("config validation warning: settings.write.buffer_size is 0, but must be " +
			"greater than 0 to allow device writing")
	}

	if c.Settings.Read.Buffer == 0 {
		logger.Warn("config validation warning: settings.read.buffer_size is 0, but must be " +
			"greater than 0 to allow device reading")
	}

	ttl, err := c.Settings.Transaction.GetTTL()
	if err != nil {
		return err
	}
	if ttl.Nanoseconds() == 0 {
		logger.Warn("config validation warning: settings.transaction.ttl is 0. transactions " +
			"will not be cached and lookups for write status will fail")
	}
	return nil
}

// NewPluginConfig creates a new PluginConfig instance which is populated from
// the configuration read in by Viper.
func NewPluginConfig() (*PluginConfig, error) {
	v := viper.New()
	setLookupInfo(v)
	return parseVersionedPluginConfig(v)
}

// setLookupInfo sets the config name, environment prefix, and search
// path(s) for a Viper instance.
func setLookupInfo(v *viper.Viper) {
	// Set the name of the config file (without the extension)
	v.SetConfigName("config")

	// Set the environment variable lookup
	v.SetEnvPrefix("plugin")
	v.AutomaticEnv()

	// If the PLUGIN_CONFIG environment variable is set, we will only
	// search for the config in that specified path, as we should expect
	// it to be there. Otherwise, we will look through a set of configuration
	// locations:
	//  * the default configuration location
	//  * a configuration directory in $HOME
	//  * the current working directory
	var configPaths []string
	configPath := os.Getenv(EnvPluginConfig)
	if configPath != "" {
		configPaths = []string{configPath}
	} else {
		configPaths = []string{
			defaultConfigPath,
			homeConfigPath,
			".",
		}
	}

	// Logging out the config paths here since it may help with debugging.
	for _, path := range configPaths {
		logger.Infof("Adding configuration path: %v", path)
		v.AddConfigPath(path)
	}
}
