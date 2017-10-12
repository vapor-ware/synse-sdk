package sdk


import (
	"fmt"
)

//
func DevicesFromConfig(dir string, deviceHandler DeviceHandler) []Device {

	protoCfg, err := ParsePrototypeConfig(dir)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}

	deviceCfg, err := ParseDeviceConfig(dir)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}

	return makeDevices(deviceCfg, protoCfg, deviceHandler)
}

//
func makeDevices(deviceConfigs []DeviceConfig, protoConfigs []PrototypeConfig, deviceHandler DeviceHandler) []Device {
	var devices []Device

	for _, dev := range deviceConfigs {
		var protoconfig PrototypeConfig
		found := false

		for _, proto := range protoConfigs {
			if proto.Type == dev.Type && proto.Model == dev.Model {
				protoconfig = proto
				found = true
				break
			}
		}

		if !found {
			// FIXME: What is the proper way to handle this?
			fmt.Printf("ERROR: Did not find the prototype for the instance!")
		}

		d := Device{
			Prototype: protoconfig,
			Instance: dev,
			Handler: deviceHandler,
		}

		fmt.Printf("  uid: %v\n", d.Uid())

		devices = append(devices, d)

	}
	return devices
}