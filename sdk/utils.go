package sdk

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
)

// makeIDString makes a compound string out of the given rack, board, and
// device identifier strings. This string should be a globally unique identifier
// for a given device.
func makeIDString(rack, board, device string) string {
	return strings.Join([]string{rack, board, device}, "-")
}

// getHandlerForDevice
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
func makeDevices(deviceConfigs []*config.DeviceConfig, protoConfigs []*config.PrototypeConfig, handlers *Handlers, devHandlers []*DeviceHandler, plugin *Plugin) ([]*Device, error) {
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

		handler, err := getHandlerForDevice(devHandlers, dev)
		if err != nil {
			return nil, err
		}

		d, err := NewDevice(
			protoconfig,
			dev,
			handler,
			plugin,
		)
		if err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
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
			return "", err
		}
	} else {
		_ = os.Remove(socket)
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
	h := md5.New()
	io.WriteString(h, protocol)
	io.WriteString(h, deviceType)
	io.WriteString(h, model)
	io.WriteString(h, protoComp)

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
