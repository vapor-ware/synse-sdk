package sdk

// PluginHandler defines the interface that a plugin implementation needs to
// fulfil in order to handle reads and writes.
//
// Within the SDK, Read and Write are are called synchronously. First, writes
// are processed, and then all of the reads are processed. This is done in
// order to support serial device. Not all protocols are serial bound, but
// we must cater to the lowest common denominator.
//
// FUTURE: A configuration option could be added to modify the read-write
// behavior to allow for parallel reads and writes.
type PluginHandler interface {

	// Read the device specified by the `ReadResource`.
	Read(Device) (ReadResource, error)

	// Write data to the specified device.
	Write(Device, *WriteData) (error)
}


// The DeviceHandler interface describes methods needed to properly parse
// protocol-specific device information.

// DeviceHandler defines the interface which a plugin implementation should
// fulfil for plugin-specific device parsing and handling.
type DeviceHandler interface {

	// GetProtocoldentifiers gets the protocol-specific pieces of information
	// that make a device instance unique. This value (or joined set of values)
	// will be used as a component when creating the device UID.
	GetProtocolIdentifiers(map[string]string) (string)

	// EnumerateDevices defines how the plugin can auto-enumerate the devices
	// it manages. This will not be relevant for all plugins, so it is optional.
	// Any plugin that does not have the ability to auto-enumerate will still
	// need to define this, but can just return the EnumerationNotSupported SDK
	// error.
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
	EnumerateDevices(map[string]interface{}) ([]DeviceConfig, error)
}
