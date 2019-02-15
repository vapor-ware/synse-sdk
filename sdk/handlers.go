package sdk

import (
	"github.com/vapor-ware/synse-sdk/sdk/config"
)

// DeviceIdentifier is a handler function that produces a string that can be used to
// identify a device deterministically. The returned string should be a composite
// from the Device's config data.
type DeviceIdentifier func(map[string]interface{}) string

// DynamicDeviceRegistrar is a handler function that takes a Plugin config's "dynamic
// registration" data and generates Device instances from it. How this is done
// is specific to the plugin/protocol.
type DynamicDeviceRegistrar func(map[string]interface{}) ([]*Device, error)

// DynamicDeviceConfigRegistrar is a handler function that takes a Plugin config's "dynamic
// registration" data and generates DeviceConfig instances from it. How this is done
// is specific to the plugin/protocol.
type DynamicDeviceConfigRegistrar func(map[string]interface{}) ([]*config.Devices, error)

// DeviceDataValidator is a handler function that takes the `Data` field of a device config
// and performs some validation on it. This allows users to provide validation on the
// plugin-specific config fields.
type DeviceDataValidator func(map[string]interface{}) error
