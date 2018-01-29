package sdk

import (
	"github.com/vapor-ware/synse-sdk/sdk/config"
)

// DeviceIdentifier gets the protocol-specific pieces of information
// that make a device instance unique. This value (or joined set of values)
// will be used as a component when creating the device UID.
type DeviceIdentifier func(map[string]string) string

// DeviceEnumerator defines how the plugin can auto-enumerate the devices
// it manages. This will not be relevant for all plugins, so it is optional.
//
// Device enumeration is when devices are not defined directly in the YAML
// configurations, but are discovered or created dynamically based on the
// capabilities of the plugin's protocol. For example, IPMI can enumerate
// devices by scanning the SDR and using those entries to make DeviceConfig
// records.
//
// Note that device auto-enumeration should only create device instance
// configurations. The device prototype configurations should still be
// defined ahead of time and packaged with the plugin - not created or
// configured at runtime.
type DeviceEnumerator func(map[string]interface{}) ([]*config.DeviceConfig, error)

// Handlers contains the user-defined handlers for a Plugin instance.
type Handlers struct {
	DeviceIdentifier
	DeviceEnumerator
}

// NewHandlers creates a new instance of Handlers.
func NewHandlers(id DeviceIdentifier, enum DeviceEnumerator) *Handlers {
	return &Handlers{
		DeviceIdentifier: id,
		DeviceEnumerator: enum,
	}
}
