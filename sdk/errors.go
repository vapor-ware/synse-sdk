package sdk


// UnsupportedCommandError is an error that can be used to designate that
// a given device does not support an operation, e.g. write. Write is required
// by the PluginHandler interface, but if a device (e.g. a temperature sensor)
// doesn't support writing, this error can be returned.
type UnsupportedCommandError struct {}

func (e *UnsupportedCommandError) Error() string {
	return "Command not supported for given device."
}
