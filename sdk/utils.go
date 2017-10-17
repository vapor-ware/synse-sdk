package sdk

// DevicesFromConfig takes the configuration directory and the plugin-defined
// DeviceHandler and generates a collection of Devices that represent all of
// the devices that are known to the plugin. These are the devices that will
// be read from, written to, and where metadata will come from.
func DevicesFromConfig(dir string, deviceHandler DeviceHandler) ([]Device, error) {

	protoCfg, err := ParsePrototypeConfig(dir)
	if err != nil {
		return nil, err
	}

	deviceCfg, err := ParseDeviceConfig(dir)
	if err != nil {
		return nil, err
	}

	return makeDevices(deviceCfg, protoCfg, deviceHandler), nil
}

// makeDevices takes the prototype and device instance configurations, parsed
// into their corresponding structs, and generates Device instances with that
// information.
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
			Logger.Warnf("Did not find prototype matching instance for %v-%v", dev.Type, dev.Model)
		}

		d := Device{
			Prototype: protoconfig,
			Instance: dev,
			Handler: deviceHandler,
		}

		Logger.Debugf("New Device: %v", d.Uid())
		devices = append(devices, d)
	}
	return devices
}