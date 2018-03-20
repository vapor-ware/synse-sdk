package config

import (
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
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

	// Create the limiter, if it is configured. If it is not configured,
	// then we will set the Limiter as nil
	var limiter *rate.Limiter
	if v.IsSet("limiter") {
		limitRate := v.GetFloat64("limiter.rate")
		limitBurst := v.GetInt("limiter.burst")

		var r rate.Limit
		// It the rate is specified as zero, we take that to mean that the
		// rate should be infinite. While zero is a valid case for the limiter,
		// it would allow no events, so for our purposes we will just use 0
		// to define "infinite".
		if limitRate == 0 {
			r = rate.Inf
		} else {
			r = rate.Limit(limitRate)
		}

		// If the burst (e.g. bucket size) is set to 0, no events are allowed.
		// We don't ever want that when a limiter is configured, so here we
		// default to have the burst size be the same as the rate limit.
		if limitBurst == 0 {
			limitBurst = int(r)
		}
		limiter = rate.NewLimiter(r, limitBurst)

	} else {
		limiter = nil
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
			Mode: v.GetString("settings.mode"),
			Read: ReadSettings{
				Enabled:  v.GetBool("settings.read.enabled"),
				Interval: v.GetString("settings.read.interval"),
				Buffer:   v.GetInt("settings.read.buffer"),
			},
			Write: WriteSettings{
				Enabled:  v.GetBool("settings.write.enabled"),
				Interval: v.GetString("settings.write.interval"),
				Buffer:   v.GetInt("settings.write.buffer"),
				Max:      v.GetInt("settings.write.max"),
			},
			Transaction: TransactionSettings{
				TTL: v.GetString("settings.transaction.ttl"),
			},
		},
		AutoEnumerate: autoEnum,
		Context:       ctx,
		Limiter:       limiter,
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
	v.SetDefault("settings.mode", "serial")
	v.SetDefault("settings.read.interval", "1s")
	v.SetDefault("settings.read.buffer", 100)
	v.SetDefault("settings.read.enabled", true)
	v.SetDefault("settings.write.interval", "1s")
	v.SetDefault("settings.write.buffer", 100)
	v.SetDefault("settings.write.max", 100) // default is same as buffer size
	v.SetDefault("settings.write.enabled", true)
	v.SetDefault("settings.transaction.ttl", "5m") // five minutes

	// auto-enumerate
	v.SetDefault("auto_enumerate", []map[string]interface{}{})

	// context
	v.SetDefault("context", map[string]interface{}{})
}
