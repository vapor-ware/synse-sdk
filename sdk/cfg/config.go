package cfg

import (
	"github.com/vapor-ware/synse-sdk/sdk/errors"
)

// ConfigComponent is an interface that all structs that define configuration
// components should implement.
//
// This interface implements a Validate function which is used by the
// SchemeValidator in order to validate each struct that makes up a configuration.
type ConfigComponent interface {
	Validate(*errors.MultiError)
}

// ConfigBase is an interface that the base configuration struct should
// implement. This allows the SchemeValidator to get the SchemeVersion
// for that given configuration.
type ConfigBase interface {
	GetSchemeVersion() (*SchemeVersion, error)
}

// ConfigContext is a structure that associates context with configuration info.
//
// The context around some bit of configuration is useful in logging/errors, as
// it lets us know which config we are talking about.
type ConfigContext struct {
	// Source is where the config came from.
	Source string

	// Config is the configuration itself. This should be a configuration struct
	// that implements ConfigBase. That is to say, the config held in this context
	// should be the root config struct for that config type. This will allow us
	// to get the scheme version of the configuration.
	Config ConfigBase
}

// NewConfigContext creates a new ConfigContext instance.
func NewConfigContext(source string, config ConfigBase) *ConfigContext {
	return &ConfigContext{
		Source: source,
		Config: config,
	}
}

// IsDeviceConfig checks whether the config in this context is a DeviceConfig.
func (ctx *ConfigContext) IsDeviceConfig() bool {
	_, ok := ctx.Config.(*DeviceConfig)
	return ok
}

// IsPluginConfig checks whether the config in the context is a PluginConfig.
func (ctx *ConfigContext) IsPluginConfig() bool {
	_, ok := ctx.Config.(*PluginConfig)
	return ok
}
