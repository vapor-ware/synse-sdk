package cfg

// ConfigComponent is an interface that all structs that define configuration
// components should implement.
//
// This interface implements a Validate function which is used by the
// SchemeValidator in order to validate each struct that makes up a configuration.
type ConfigComponent interface {
	Validate() error
}

// ConfigContext is a structure that associates context with configuration info.
//
// The context around some bit of configuration is useful in logging/errors, as
// it lets us know which config we are talking about.
type ConfigContext struct {
	// Source is where the config came from.
	Source string

	// Config is the configuration itself.
	Config interface{}
}

// NewConfigContext creates a new ConfigContext instance.
func NewConfigContext(source string, config interface{}) *ConfigContext {
	return &ConfigContext{
		Source: source,
		Config: config,
	}
}

// IsDeviceConfig checks whether the config in this context
// is a DeviceConfig.
func (ctx *ConfigContext) IsDeviceConfig() bool {
	_, ok := ctx.Config.(DeviceConfig)
	return ok
}
