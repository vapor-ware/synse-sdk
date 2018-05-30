package sdk

import (
	"github.com/vapor-ware/synse-sdk/sdk/cfg"
)

// FIXME: there has to be a better way to set the config globally..
// for now, this is fine since it gets things moving, but it needs
// to be improved/fixed before v1.0 is considered done.

var DeviceConfig cfg.DeviceConfig
var PluginConfig cfg.PluginConfig

// IsConfigured is a stand-in for a means by which we will determine whether
// the plugin has been globally configured or not.
func IsConfigured() bool {
	return true
}
