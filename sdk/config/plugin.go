package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

const (
	defaultConfigPath = "/etc/synse/plugin"
	homeConfigPath    = "$HOME/.synse/plugin"
)

const (
	EnvPluginConfig = "PLUGIN_CONFIG"
)

type PluginConfig struct {
	Name          string
	Version       string
	Debug         bool
	Settings      Settings
	Network       NetworkSettings
	AutoEnumerate []map[string]interface{}
	Context       map[string]interface{}
}

type NetworkSettings struct {
	Type    string
	Address string
}

type Settings struct {
	LoopDelay   int
	Read        ReadSettings
	Write       WriteSettings
	Transaction TransactionSettings
}

type ReadSettings struct {
	BufferSize int
}

type WriteSettings struct {
	BufferSize int
	PerLoop    int
}

type TransactionSettings struct {
	TTL int
}

// Validate checks the PluginConfig instance to make sure it has all of
// the required fields populated.
func (c *PluginConfig) Validate() error {
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
	return nil
}

func NewPluginConfig() (*PluginConfig, error) {
	v := viper.New()
	setLookupInfo(v)
	setDefaults(v)

	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	autoEnum, err := toSliceStringMapI(v.Get("auto_enumerate"))
	if err != nil {
		return nil, err
	}

	ctx, err := toStringMapI(v.Get("context"))
	if err != nil {
		return nil, err
	}

	p := &PluginConfig{
		Name: v.GetString("name"),
		Version: v.GetString("version"),
		Network: NetworkSettings{
			Type: v.GetString("network.type"),
			Address: v.GetString("network.address"),
		},
		Settings: Settings{
			LoopDelay: v.GetInt("settings.loop_delay"),
			Read: ReadSettings{
				BufferSize: v.GetInt("settings.read.buffer_size"),
			},
			Write: WriteSettings{
				BufferSize: v.GetInt("settings.write.buffer_size"),
				PerLoop: v.GetInt("settings.write.per_loop"),
			},
			Transaction: TransactionSettings{
				TTL: v.GetInt("settings.transaction.ttl"),
			},
		},
		AutoEnumerate: autoEnum,
		Context: ctx,
	}

	err = p.Validate()
	if err != nil {
		return nil, err
	}
	return p, nil
}

func setDefaults(v *viper.Viper) {
	// the "name", "version" and "network" fields are required, so they should
	// not have any default values.

	v.SetDefault("debug", false)

	// settings
	v.SetDefault("settings.loop_delay", 0)
	v.SetDefault("settings.read.buffer_size", 100)
	v.SetDefault("settings.write.buffer_size", 100)
	v.SetDefault("settings.write.per_loop", 5)
	v.SetDefault("settings.transaction.ttl", 60*5) // five minutes

	// auto-enumerate
	v.SetDefault("auto_enumerate", []map[string]interface{}{})

	// context
	v.SetDefault("context", map[string]interface{}{})
}

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
	configPath := os.Getenv(EnvPluginConfig)
	if configPath != "" {
		v.AddConfigPath(configPath)
	} else {
		v.AddConfigPath(defaultConfigPath)
		v.AddConfigPath(homeConfigPath)
		v.AddConfigPath(".")
	}
}
