package config

import (
	"github.com/spf13/viper"
)

type v1PluginConfigHandler struct{}

func (h *v1PluginConfigHandler) processPluginConfig(v *viper.Viper) (*PluginConfig, error) {

	// Set any default values for the v1 configuration
	setV1Defaults(v)

	// Cast the "auto_enumerate" field
	autoEnum, err := toSliceStringMapI(v.Get("auto_enumerate"))
	if err != nil {
		return nil, err
	}

	// Cast the "context" field
	ctx, err := toStringMapI(v.Get("context"))
	if err != nil {
		return nil, err
	}

	// Create a new PluginConfig instance
	p := &PluginConfig{
		Name:    v.GetString("name"),
		Version: v.GetString("version"),
		Debug:   v.GetBool("debug"),
		Network: NetworkSettings{
			Type:    v.GetString("network.type"),
			Address: v.GetString("network.address"),
		},
		Settings: Settings{
			LoopDelay: v.GetInt("settings.loop_delay"),
			Read: ReadSettings{
				BufferSize: v.GetInt("settings.read.buffer_size"),
			},
			Write: WriteSettings{
				BufferSize: v.GetInt("settings.write.buffer_size"),
				PerLoop:    v.GetInt("settings.write.per_loop"),
			},
			Transaction: TransactionSettings{
				TTL: v.GetInt("settings.transaction.ttl"),
			},
		},
		AutoEnumerate: autoEnum,
		Context:       ctx,
	}

	// Validate that the PluginConfig has all of its required fields
	// populated.
	err = p.Validate()
	if err != nil {
		return nil, err
	}
	return p, nil
}

// setV1Defaults sets default v1 configuration values for a Viper instance.
func setV1Defaults(v *viper.Viper) {
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
