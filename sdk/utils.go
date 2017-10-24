package sdk


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
			break
		}

		d := Device{
			Prototype: protoconfig,
			Instance: dev,
			Handler: deviceHandler,
		}

		Logger.Debugf("New Device: %v", d.UID())
		devices = append(devices, d)
	}
	return devices
}