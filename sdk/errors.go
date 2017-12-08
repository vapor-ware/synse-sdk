package sdk

// UnsupportedCommandError is an error that can be used to designate that
// a given device does not support an operation, e.g. write. Write is required
// by the PluginHandler interface, but if a device (e.g. a temperature sensor)
// doesn't support writing, this error can be returned.
type UnsupportedCommandError struct{}

func (e *UnsupportedCommandError) Error() string {
	return "Command not supported for given device."
}

// EnumerationNotSupported is an error that should be returned when defining a
// plugin's DeviceHandler interface and it does not support the EnumerateDevices
// function.
type EnumerationNotSupported struct{}

func (e *EnumerationNotSupported) Error() string {
	return "This plugin does not support device auto-enumeration."
}
