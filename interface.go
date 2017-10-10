package sdk

import (
	synse "github.com/vapor-ware/synse-server-grpc/go"
)

// Every plugin that is written using this SDK must fulfil this interface. These
// functions are what the Read-Write loop will end up calling when the corresponding
// internal API calls are made.
//
// The plugin only needs to specify the behavior for the read and write commands
// here. The transaction check and metainfo commands are fulfilled the same way
// for each plugin, so those commands are provided by this SDK.
type PluginHandler interface {
	Read(in Device) (ReadResource, error)
	Write(in Device, data []string) (*synse.TransactionId, error)
}


// The DeviceHandler interface describes methods needed to properly parse
// protocol-specific device information.
type DeviceHandler interface {
	// Get the protocol-specific pieces of information that make a device instance
	// unique. This value (or joined values) will be used as a component when creating
	// the device UID.
	GetProtocolIdentifiers(data map[string]string) (string)
}
