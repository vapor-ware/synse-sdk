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
)

// makeIDString makes a compound string out of the given rack, board, and
// device identifier strings. This string should be a globally unique identifier
// for a given device.
func makeIDString(rack, board, device string) string {
	return strings.Join([]string{rack, board, device}, "-")
}

// getHandlerForDevice gets the DeviceHandler for the device, based on its
// Type and Model.
func getHandlerForDevice(handlers []*DeviceHandler, device *config.DeviceConfig) (*DeviceHandler, error) {
	for _, h := range handlers {
		if device.Type == h.Type && device.Model == h.Model {
			return h, nil
		}
	}
	return nil, fmt.Errorf("no handler found for device %#v", device)
}

// makeDevices takes the prototype and device instance configurations, parsed
// into their corresponding structs, and generates Device instances with that
// information.
func makeDevices(deviceConfigs []*config.DeviceConfig, protoConfigs []*config.PrototypeConfig, plugin *Plugin) ([]*Device, error) {
	logger.Debugf("makeDevices start")

	var devices []*Device
	for _, dev := range deviceConfigs {
		var protoconfig *config.PrototypeConfig
		found := false

		for _, proto := range protoConfigs {
			if proto.Type == dev.Type && proto.Model == dev.Model {
				protoconfig = proto
				found = true
				break
			}
		}

		if !found {
			logger.Warnf("Did not find prototype matching instance for %v-%v", dev.Type, dev.Model)
			break
		}
		logger.Debugf("Found prototype matching instance config for %v %v", dev.Type, dev.Model)

		handler, err := getHandlerForDevice(plugin.deviceHandlers, dev)
		if err != nil {
			logger.Errorf("found no handler for device %v: %v", dev, err)
			return nil, err
		}

		d, err := NewDevice(
			protoconfig,
			dev,
			handler,
			plugin,
		)
		if err != nil {
			logger.Errorf("failed to create new device: %v", err)
			return nil, err
		}
		devices = append(devices, d)
	}

	logger.Debugf("finished making devices: %v", devices)
	return devices, nil
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
func GetCurrentTime() string {
	return time.Now().Format(time.RFC3339Nano)
}
