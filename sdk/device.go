// Synse SDK
// Copyright (c) 2019 Vapor IO
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package sdk

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/imdario/mergo"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/utils"
	synse "github.com/vapor-ware/synse-server-grpc/go"
)

// TODO (etd): consider not exporting the Device fields. The reason being that
//  while a plugin may need to interact with a device, it should never really
//  be modifying device data once it has been loaded (I don't think..)

// Device is a single physical or virtual device which the Plugin manages.
//
// It defines all of the information known about the device, which typically
// comes from configuration file. A Device's supported actions are determined
// by the DeviceHandler which it is configured to use.
type Device struct {
	// Type is the type of the device. This is largely metadata that can
	// be used upstream to categorize the device.
	Type string

	// Metadata is arbitrary metadata that is associated with the device.
	Metadata map[string]string

	// Info is a human-readable string that provides a summary of what
	// the device is or what it does.
	Info string

	// Tags is the set of tags that are associated with the device.
	Tags []*Tag

	// Data contains any plugin-specific configuration data for the device.
	Data map[string]interface{}

	// Handler is the name of the device's handler.
	Handler string

	// SortIndex is an optional 1-based index sort ordinal that is used upstream
	// (by Synse Server) to sort the device in Scan output. This can be set to
	// give devices custom ordering. If not set, they will be sorted based on
	// default sorting parameters.
	SortIndex int32

	// Alias is a human-readable alias for the device which can be used to
	// to reference it as well.
	Alias string

	// ScalingFactor is an optional value by which to scale the device readings.
	// If defined, the reading values for the device will be multiplied by this
	// value.
	//
	// This value should resolve to a numeric. Negative values and fractional values
	// are supported. This can be the value itself, e.g. "0.01", or a mathematical
	// representation of the value, e.g. "1e-2".
	ScalingFactor string

	// System defines the System of Measure for the device. It is the default
	// system of measure (imperial, metric) for the device's reading data. This
	// is not required in all cases, as the DeviceHandler may specify the system
	// as well. Generally, this should only be set for devices using generalized
	// device handlers which do not define a system.
	System string

	// WriteTimeout defines the time within which a write action (transaction)
	// will remain valid for this device.
	WriteTimeout time.Duration

	// Output is the name of the Output that this device instance will use. This
	// is not needed for all devices/plugins, as many DeviceHandlers will already
	// know which output to use. This field is used in cases of generalized plugins,
	// such as Modbus-IP, where a generalized handler will need to map something
	// (like a set of registers) to a reading output.
	Output string

	// id is the unique ID for the device.
	id string

	// idName is the generated name of the device based on its components which
	// is ultimately used to generate its ID.
	idName string

	// handler is a pointer to the actual DeviceHandler for the device. This is
	// populated via the SDK on device loading and parsing and uses the Handler
	// field to match the name of the handler to the actual instance.
	handler *DeviceHandler
}

// NewDeviceFromConfig creates a new instance of a Device from its device prototype
// and device instance configuration.
//
// These configuration components are loaded from config file.
func NewDeviceFromConfig(proto *config.DeviceProto, instance *config.DeviceInstance) (*Device, error) {
	// Define variable for the Device fields that can be inherited from the
	// device prototype configuration.
	var (
		data         map[string]interface{}
		tags         []string
		handler      string
		system       string
		deviceType   string
		writeTimeout time.Duration
	)

	// If inheritance is enabled, use the prototype defined value as the base.
	if !instance.DisableInheritance {
		data = proto.Data
		tags = proto.Tags
		handler = proto.Handler
		system = proto.System
		deviceType = proto.Type
		writeTimeout = proto.WriteTimeout
	}

	// If the instance also defines the same variable, we either need to merge
	// the values or overwrite them.

	// Merge instance data.
	if err := mergo.Map(&data, instance.Data, mergo.WithOverride, mergo.WithAppendSlice); err != nil {
		log.WithField("error", err).Error("[device] failed merging device instance config")
		return nil, err
	}

	// Merge tags. It is okay if the same tag is defined more than once, (e.g.
	// no need to error), but we do ultimately just want the set of tags.
	tags = append(tags, instance.Tags...)
	var deviceTags []*Tag
	encountered := map[string]struct{}{}
	for _, t := range tags {
		if _, ok := encountered[t]; !ok {
			encountered[t] = struct{}{}
			tag, err := NewTag(t)
			if err != nil {
				return nil, err
			}
			deviceTags = append(deviceTags, tag)
		}
	}

	// Override handler, if set.
	if instance.Handler != "" {
		handler = instance.Handler
	}

	// Override system, if set.
	if instance.System != "" {
		system = instance.System
	}

	// Override type, if set.
	if instance.Type != "" {
		deviceType = instance.Type
	}
	// We require devices to have a type; error if there is none set.
	if deviceType == "" {
		// fixme: err message
		return nil, fmt.Errorf("device requires type")
	}

	// Override write timeout, if set.
	if instance.WriteTimeout != 0 {
		writeTimeout = instance.WriteTimeout
	}

	// TODO: generate the device alias

	return &Device{
		Type:          deviceType,
		Tags:          deviceTags,
		Data:          data,
		Handler:       handler,
		System:        system,
		Metadata:      proto.Metadata,
		Info:          instance.Info,
		SortIndex:     instance.SortIndex,
		ScalingFactor: instance.ScalingFactor,
		WriteTimeout:  writeTimeout,
		Output:        instance.Output,
	}, nil
}

// JSON encodes the device as JSON. This can be useful for logging and debugging.
func (device *Device) JSON() (string, error) {
	bytes, err := json.Marshal(device)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// GetHandler gets the DeviceHandler of the device.
func (device *Device) GetHandler() *DeviceHandler {
	return device.handler
}

// GetID gets the unique ID of the device.
func (device *Device) GetID() string {
	return device.id
}

// Read performs the read action for the device, as set by its DeviceHandler.
//
// If reading is not supported on the device, an UnsupportedCommandError is
// returned.
// FIXME: should we update the unsupported command error to be more descriptive?
func (device *Device) Read() (*ReadContext, error) {
	// Bulk read is handled elsewhere.
	// Device may only support bulk read.
	if device == nil {
		return nil, fmt.Errorf("device is nil")
	}
	if device.handler == nil {
		return nil, fmt.Errorf("device.handler is nil")
	}
	if device.handler.Read != nil {
		readings, err := device.handler.Read(device)
		if err != nil {
			return nil, err
		}

		return NewReadContext(device, readings), nil
	}
	return nil, &errors.UnsupportedCommandError{}
}

// Write performs the write action for the device, as set by its DeviceHandler.
//
// If writing is not supported on the device, an UnsupportedCommandError is
// returned.
// FIXME: should we update the unsupported command error to be more descriptive?
func (device *Device) Write(data *WriteData) error {
	if device.IsWritable() {
		return device.handler.Write(device, data)
	}
	return &errors.UnsupportedCommandError{}
}

// IsReadable checks if the Device is readable based on the presence/absence
// of a Read/BulkRead action defined in its DeviceHandler.
func (device *Device) IsReadable() bool {
	if device == nil {
		return false
	}
	return device.handler.Read != nil || device.handler.BulkRead != nil || device.handler.Listen != nil
}

// IsWritable checks if the Device is writable based on the presence/absence
// of a Write action defined in its DeviceHandler.
func (device *Device) IsWritable() bool {
	if device == nil {
		return false
	}
	return device.handler.Write != nil
}

// fixme: device ID generation

//// ID generates the deterministic ID for the Device using its config values.
//func (device *Device) ID() string {
//	if device.id == "" {
//		protocolComp := ctx.deviceIdentifier(device.Data)
//		device.id = utils.NewUID(device.Plugin, device.Kind, protocolComp)
//	}
//	return device.id
//}
//
//// GUID generates a globally unique ID string by creating a composite
//// string from the rack, board, and device UID.
//func (device *Device) GUID() string {
//	return utils.MakeIDString( // fixme
//		"", //device.Location.Rack,
//		"", //device.Location.Board,
//		device.ID(),
//	)
//}

// encode translates the Device to the corresponding gRPC Device message.
func (device *Device) encode() *synse.V3Device {
	var tags []*synse.V3Tag
	for _, t := range device.Tags {
		tags = append(tags, t.Encode())
	}

	return &synse.V3Device{
		Timestamp: utils.GetCurrentTime(),
		Id:        device.id,
		Type:      device.Type,
		Info:      device.Info,
		Metadata:  device.Metadata,
		SortIndex: device.SortIndex,
		Tags:      tags,
		// todo:  capabilities, outputs
	}
}

//// ValidateDeviceConfigData validates the `Data` field(s) of a Device Config to
//// ensure that they are correct. The `Data` fields are plugin-specific, so its
//// up to the user to provide us with a validation function.
//func (config *DeviceConfig) ValidateDeviceConfigData(validator func(map[string]interface{}) error) *errors.MultiError {
//	multiErr := errors.NewMultiError("device config 'data' field validation")
//
//	for _, device := range config.Devices {
//		// Verify that the DeviceKind Instances' `Data` field is correct
//		for _, instance := range device.Instances {
//			err := validator(instance.Data)
//			if err != nil {
//				multiErr.Add(err)
//			}
//			// Instance Outputs can have their own data too. Verify instance
//			// output data.
//			for _, output := range instance.Outputs {
//				err := validator(output.Data)
//				if err != nil {
//					multiErr.Add(err)
//				}
//			}
//		}
//
//		// Device kind outputs can have their own data too. Verify the
//		// device kind output data.
//		for _, output := range device.Outputs {
//			err := validator(output.Data)
//			if err != nil {
//				multiErr.Add(err)
//			}
//		}
//	}
//	return multiErr
//}
