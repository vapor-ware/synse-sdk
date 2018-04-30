package sdk

import (
	"github.com/vapor-ware/synse-sdk/sdk/config"
)

// DeviceIdentifier gets the protocol-specific pieces of information
// that make a device instance unique. This value (or joined set of values)
// will be used as a component when creating the device UID which shows up
// in a scan.
type DeviceIdentifier func(map[string]string) string

// DeviceEnumerator defines how the plugin can auto-enumerate the devices
// it manages. This will not be relevant for all plugins, so it is an
// optional handler.
//
// Device enumeration is when devices are not defined directly in the YAML
// instance configurations, but are discovered or created dynamically based
// on the capabilities of the plugin's protocol. For example, IPMI can enumerate
// devices by scanning the SDR and using those entries to make DeviceConfig
// records.
//
// Note that device enumeration should only create device instance
// configurations. The device prototype configurations should still be
// defined ahead of time and packaged with the plugin - not created or
// configured at runtime.
type DeviceEnumerator func(map[string]interface{}) ([]*config.DeviceConfig, error)

// Handlers is a struct that holds plugin-specific function handlers for
// performing SDK actions that are specific to the plugin. There are two such handlers.
type Handlers struct {
	DeviceIdentifier
	DeviceEnumerator
}

// NewHandlers creates a new instance of Handlers.
func NewHandlers(id DeviceIdentifier, enum DeviceEnumerator) (*Handlers, error) {
	if id == nil {
		return nil, invalidArgumentErr("id parameter must not be nil")
	}
	return &Handlers{
		DeviceIdentifier: id,
		DeviceEnumerator: enum,
	}, nil
}
