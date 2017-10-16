package sdk


// Every plugin that is written using this SDK must fulfil this interface. These
// functions are what the Read-Write loop will end up calling when the corresponding
// internal API calls are made.
//
// The plugin only needs to specify the behavior for the read and write commands
// here. The transaction check and metainfo commands are fulfilled the same way
// for each plugin, so those commands are provided by this SDK.
type PluginHandler interface {

	// within the sdk, read and write are called synchronously. this is done in
	// order to support serial devices (e.g. devices that communicate over a serial
	// bus). not all protocols are serial bound, but here we must cater to the lowest
	// common denominator.

	// TODO (etd) - possibly add in a configuration option that would process reads
	// in parallel and writes in parallel?

	Read(Device) (ReadResource, error)
	Write(Device, *WriteData) (error)
}


// The DeviceHandler interface describes methods needed to properly parse
// protocol-specific device information.
type DeviceHandler interface {
	// Get the protocol-specific pieces of information that make a device instance
	// unique. This value (or joined values) will be used as a component when creating
	// the device UID.
	GetProtocolIdentifiers(map[string]string) (string)
}
