package sdk

import (
	"crypto/md5" // #nosec
	"fmt"
	"io"
	"strings"
	"time"
)

// makeIDString makes a compound string out of the given rack, board, and
// device identifier strings. This string should be a globally unique identifier
// for a given device.
func makeIDString(rack, board, device string) string {
	return strings.Join([]string{rack, board, device}, "-")
}

// getTypeByName gets the output type with the given name. If an output type does
// not exist with the given name, an error is returned.
func getTypeByName(name string) (*OutputType, error) {
	t, ok := ctx.outputTypes[name]
	if !ok {
		return nil, fmt.Errorf("no output type with name '%s' found", name)
	}
	return t, nil
}

// newUID creates a new unique identifier for a Device. This id should be
// deterministic because it is a hash of various Device configuration components.
// A device's config should be unique, so the hash should be unique.
//
// These device IDs are not guaranteed to be globally unique, but they should
// be unique to the board they reside on.
func newUID(components ...string) string {
	h := md5.New() // nolint: gas
	for _, component := range components {
		io.WriteString(h, component) // nolint: errcheck
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// filterDevices returns a list of Devices (a subset of the deviceMap) which
// match the specified filter(s) in the given filter string.
func filterDevices(filter string) ([]*Device, error) { // nolint: gocyclo
	filters := strings.Split(filter, ",")

	var devices []*Device
	for _, d := range ctx.devices {
		devices = append(devices, d)
	}

	for _, f := range filters {
		pair := strings.Split(f, "=")
		if len(pair) != 2 {
			return nil, fmt.Errorf("incorrect filter string: %s", filter)
		}
		k, v := pair[0], pair[1]

		var isValid func(d *Device) bool
		switch k {
		case "kind":
			isValid = func(d *Device) bool { return d.Kind == v || v == "*" }
		case "type":
			isValid = func(d *Device) bool { return d.GetType() == v || v == "*" }
		default:
			return nil, fmt.Errorf("unsupported filter key. expect 'kind' but got %s", k)
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
