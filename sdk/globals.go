package sdk

import (
	"github.com/vapor-ware/synse-sdk/sdk/config"
)

// FIXME: there has to be a better way to set the config globally..
// for now, this is fine since it gets things moving, but it needs
// to be improved/fixed before v1.0 is considered done.

var DeviceConfig config.DeviceConfig
var PluginConfig config.PluginConfig