package sdk

import (
	"crypto/md5" // #nosec
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
	"github.com/vapor-ware/synse-sdk/sdk/types"
)

// makeIDString makes a compound string out of the given rack, board, and
// device identifier strings. This string should be a globally unique identifier
// for a given device.
func makeIDString(rack, board, device string) string {
	return strings.Join([]string{rack, board, device}, "-")
}

// getHandlerForDevice gets the DeviceHandler for the device, based on its
// Type and Model.
func getHandlerForDevice(handlerName string) (*DeviceHandler, error) {
	for _, handler := range deviceHandlers {
		if handler.Name == handlerName {
			return handler, nil
		}
	}
	return nil, fmt.Errorf("no handler found with name: %s", handlerName)
}

// makeDevices
func makeDevices(config *config.DeviceConfig, plugin *Plugin) ([]*Device, error) {
	logger.Debugf("makeDevices start")

	// the list of devices we made
	var devices []*Device

	// the DeviceConfig we get here should be the unified config.
	for _, kind := range config.Devices {
		for _, instance := range kind.Instances {

			// create the outputs for the instance.
			var instanceOutputs []*Output
			for _, o := range instance.Outputs {
				output, err := NewOutputFromConfig(o)
				if err != nil {
					return nil, err
				}
				instanceOutputs = append(instanceOutputs, output)
			}

			if instance.InheritKindOutputs {
				for _, o := range kind.Outputs {
					output, err := NewOutputFromConfig(o)
					if err != nil {
						return nil, err
					}
					instanceOutputs = append(instanceOutputs, output)
				}
			}

			// Get the location
			location, err := config.GetLocation(instance.Location)
			if err != nil {
				return nil, err
			}

			// If a specific handlerName is set in the config, we will use that as the
			// definitive handler. Otherwise, use the kind.
			handlerName := kind.Name
			if kind.HandlerName != "" {
				handlerName = kind.HandlerName
			}
			if instance.HandlerName != "" {
				handlerName = instance.HandlerName
			}

			// Get the DeviceHandler
			handler, err := getHandlerForDevice(handlerName)
			if err != nil {
				return nil, err
			}

			device := &Device{
				// The name of the device kind. This is essentially the identifier
				// for the device type.
				Kind: kind.Name,

				// Any metadata associated with the device kind.
				Metadata: kind.Metadata,

				// The name of the plugin.
				Plugin: metainfo.Name,

				// Device-level information. This is not reading output level info.
				Info: instance.Info,

				// The location of the device.
				Location: location,

				// Any data associated with the device instance.
				Data: instance.Data,

				// The outputs supported by the device. A device output may
				// supply more info, like Data, Info, Type, etc, so that should
				// be accounted for when doing readings/writing stuff..
				Outputs: instanceOutputs,

				// The read/write handler for the device. Handlers should be registered globally.
				Handler: handler,
			}

			devices = append(devices, device)

		}

	}

	//var devices []*Device
	//for _, dev := range deviceConfigs {
	//	var protoconfig *config.PrototypeConfig
	//	found := false
	//
	//	for _, proto := range protoConfigs {
	//		if proto.Type == dev.Type && proto.Model == dev.Model {
	//			protoconfig = proto
	//			found = true
	//			break
	//		}
	//	}
	//
	//	if !found {
	//		logger.Warnf("Did not find prototype matching instance for %v-%v", dev.Type, dev.Model)
	//		continue
	//	}
	//	logger.Debugf("Found prototype matching instance config for %v %v", dev.Type, dev.Model)
	//
	//	handler, err := getHandlerForDevice(plugin.deviceHandlers, dev)
	//	if err != nil {
	//		logger.Errorf("found no handler for device %v: %v", dev, err)
	//		return nil, err
	//	}
	//
	//	d, err := NewDevice(
	//		protoconfig,
	//		dev,
	//		handler,
	//		plugin,
	//	)
	//	if err != nil {
	//		logger.Errorf("failed to create new device: %v", err)
	//		return nil, err
	//	}
	//	devices = append(devices, d)
	//}
	//
	//logger.Debugf("finished making devices: %v", devices)
	return devices, nil
}

// TODO -- implement
func getTypeByName(name string) (*types.ReadingType, error) {
	return nil, nil
}

// setupSocket is used to make sure the path for unix socket used for gRPC communication
// is set up and accessible locally. Creates the directory for the socket. Returns the
// directoryName and err.
func setupSocket(name string) (string, error) {
	socket := fmt.Sprintf("%s/%s", sockPath, name)

	_, err := os.Stat(sockPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(sockPath, os.ModePerm); err != nil {
				return "", err
			}
		} else {
			logger.Errorf("failed to create socket path %v: %v", sockPath, err)
			return "", err
		}
	} else {
		err = os.Remove(socket)
		if err != nil {
			if !os.IsNotExist(err) {
				logger.Errorf("failed to remove existing socket %v: %v", socket, err)
				return "", err
			}
		}
	}
	return socket, nil
}

// newUID creates a new unique identifier for a device. The device id is
// deterministic because it is created as a hash of various components that
// make up the device's configuration. By definition, each device will have
// a (slightly) different configuration (otherwise they would just be the same
// devices).
//
// These device IDs are not guaranteed to be globally unique, but they should
// be unique to the board they reside on.
func newUID(protocol, deviceType, model, protoComp string) string {
	h := md5.New()                // nolint: gas
	io.WriteString(h, protocol)   // nolint: errcheck
	io.WriteString(h, deviceType) // nolint: errcheck
	io.WriteString(h, model)      // nolint: errcheck
	io.WriteString(h, protoComp)  // nolint: errcheck

	return fmt.Sprintf("%x", h.Sum(nil))
}

// filterDevices returns a list of Devices (a subset of the deviceMap) which
// match the specified filter(s) in the given filter string.
func filterDevices(filter string) ([]*Device, error) {
	filters := strings.Split(filter, ",")

	var devices []*Device
	for _, d := range deviceMap {
		devices = append(devices, d)
	}

	for _, f := range filters {
		pair := strings.Split(f, "=")
		if len(pair) != 2 {
			return nil, fmt.Errorf("incorrect filter string: %s", filter)
		}
		k, v := pair[0], pair[1]

		var isValid func(d *Device) bool
		if k == "type" {
			isValid = func(d *Device) bool { return d.Type == v || v == "*" }
		} else if k == "model" {
			isValid = func(d *Device) bool { return d.Model == v || v == "*" }
		} else {
			return nil, fmt.Errorf("unsupported filter key. expect 'type' or 'string' but got %s", k)
		}

		i := 0
		for _, d := range devices {
			if isValid(d) {
				devices[i] = d
				i++
			}
		}
		devices = devices[:i]
	}
	return devices, nil
}

// GetCurrentTime return the current time (time.Now()) as a string formatted
// with the RFC3339Nano layout. This should be the format of all timestamps
// returned by the SDK.
//
// The SDK uses this function to generate all of its timestamps. It is highly
// recommended that plugins use this as well for timestamp generation.
func GetCurrentTime() string {
	return time.Now().Format(time.RFC3339Nano)
}
