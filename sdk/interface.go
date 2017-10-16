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
}
