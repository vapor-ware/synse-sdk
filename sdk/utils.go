package sdk

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/vapor-ware/synse-sdk/sdk/config"
)

// makeIDString makes a compound string out of the given rack, board, and
// device identifier strings. This string should be a globally unique identifier
// for a given device.
func makeIDString(rack, board, device string) string {
	return strings.Join([]string{rack, board, device}, "-")
}

// makeDevices takes the prototype and device instance configurations, parsed
// into their corresponding structs, and generates Device instances with that
// information.
func makeDevices(deviceConfigs []*config.DeviceConfig, protoConfigs []*config.PrototypeConfig, deviceHandler DeviceHandler) []*Device {
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
			Logger.Warnf("Did not find prototype matching instance for %v-%v", dev.Type, dev.Model)
			break
		}

		d := Device{
			Prototype: protoconfig,
			Instance:  dev,
			Handler:   deviceHandler,
		}
		devices = append(devices, &d)
	}
	return devices
}

// setupSocket is used to make sure the unix socket used for gRPC communication
// is set up and accessible locally.
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
