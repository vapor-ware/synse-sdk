package sdk

import (
	"fmt"
	"sort"

	"github.com/vapor-ware/synse-sdk/sdk/config"
)

// defaultDeviceIdentifier is the default implementation that fulfils the DeviceIdentifier
// type for a plugin context.
//
// This implementation creates a string by joining all values found in the provided
// device data map. Non-string values in the map are cast to a string.
func defaultDeviceIdentifier(data map[string]interface{}) string {
	var identifier string

	// To ensure that we get the same identifier reliably, we want to make sure
	// we append the components reliably, so we will sort the keys.
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		// Instead of implementing our own type checking and casting, just
		// use Sprint. Note that this may be meaningless for complex types.
		// TODO: write tests/check to see how this behaves with things like
		//  maps/lists. I have a feeling that maps will not produce a deterministic
		//  string because they are unordered, so we may need custom handling for that.
		identifier += fmt.Sprint(data[key])
	}
	return identifier
}

// defaultDynamicDeviceRegistration is the default implementation that fulfils the
// DynamicDeviceRegistrar type for a plugin context.
//
// This implementation simply returns an empty slice. A plugin will not do any dynamic
// registration by default.
func defaultDynamicDeviceRegistration(_ map[string]interface{}) ([]*Device, error) {
	return []*Device{}, nil
}

// defaultDynamicDeviceConfigRegistration is the default implementation that fulfils the
// DynamicDeviceConfigRegistrar type for a plugin context.
//
// This implementation simply returns an empty slice. A plugin will not do any dynamic
// registration by default.
func defaultDynamicDeviceConfigRegistration(_ map[string]interface{}) ([]*config.DeviceConfig, error) {
	return []*config.DeviceConfig{}, nil
}

// defaultDeviceDataValidator is the default implementation that fulfils the
// DeviceDataValidator type for a plugin context.
//
// This implementation simply returns nil. By default, this will not do any custom
// validation.
func defaultDeviceDataValidator(_ map[string]interface{}) error {
	return nil
}
