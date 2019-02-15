package sdk

import (
	"fmt"
	"reflect"
	"sort"

	// TODO: "config" is in the package namespace.. we'll need to clean
	//  that up so we don't need to alias the import
	cfg "github.com/vapor-ware/synse-sdk/sdk/config"

	log "github.com/Sirupsen/logrus"
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
		value := data[key]

		// Check if the value is a map. If so, ignore it. Since maps are
		// not ordered, we cannot use them to create a stable device id.
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Map {
			log.Debug("[sdk] default device identifier - data value is map; skipping")
			continue
		}

		// Instead of implementing our own type checking and casting, just
		// use Sprint. Note that this may be meaningless for complex types.
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
func defaultDynamicDeviceConfigRegistration(_ map[string]interface{}) ([]*cfg.Devices, error) {
	return []*cfg.Devices{}, nil
}

// defaultDeviceDataValidator is the default implementation that fulfils the
// DeviceDataValidator type for a plugin context.
//
// This implementation simply returns nil. By default, this will not do any custom
// validation.
func defaultDeviceDataValidator(_ map[string]interface{}) error {
	return nil
}
