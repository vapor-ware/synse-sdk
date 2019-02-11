package sdk

import (
	"crypto/md5" // #nosec
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/policies"
)

// GetCurrentTime return the current time (time.Now()), with location set to UTC,
// as a string formatted with the RFC3339 layout. This should be the format
// of all timestamps returned by the SDK.
//
// The SDK uses this function to generate all of its timestamps. It is highly
// recommended that plugins use this as well for timestamp generation.
func GetCurrentTime() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// ParseRFC3339 parses a timestamp string in RFC3339 format into a Time struct.
// If it is given an empty string, it will return the zero-value for a Time
// instance. You can check if it is a zero time with the Time's `IsZero` method.
func ParseRFC3339(timestamp string) (t time.Time, err error) {
	if timestamp == "" {
		return
	}
	t, err = time.Parse(time.RFC3339, timestamp)
	if err != nil {
		log.WithField(
			"timestamp", timestamp,
		).Error("[sdk] failed to parse timestamp from RFC3339 format")
	}
	return
}

// GetTypeByName gets the output type with the given name from the collection of
// output types registered with the SDK for the plugin. If an output type with the
// given name does not exist, an error is returned.
func GetTypeByName(name string) (*OutputType, error) {
	t, ok := ctx.outputTypes[name]
	if !ok {
		return nil, fmt.Errorf("no output type with name '%s' found", name)
	}
	return t, nil
}

// ConvertToFloat64 converts value to a float64 or errors out.
func ConvertToFloat64(value interface{}) (result float64, err error) { // nolint: gocyclo
	switch t := value.(type) {
	case float64:
		result = t
	case float32:
		result = float64(t)
	case int64:
		result = float64(t)
	case int32:
		result = float64(t)
	case int16:
		result = float64(t)
	case int8:
		result = float64(t)
	case int:
		result = float64(t)
	case uint64:
		result = float64(t)
	case uint32:
		result = float64(t)
	case uint16:
		result = float64(t)
	case uint8:
		result = float64(t)
	case uint:
		result = float64(t)
	case string:
		result, err = strconv.ParseFloat(t, 64)
	default:
		err = fmt.Errorf("Unable to convert value %v, type %T to float64", value, value)
	}
	return result, err
}

// makeIDString makes a compound string out of the given rack, board, and
// device identifier strings. This string should be a globally unique identifier
// for a given device.
func makeIDString(rack, board, device string) string {
	return strings.Join([]string{rack, board, device}, "-")
}

// newUID creates a new unique identifier for a Device. This id should be
// deterministic because it is a hash of various Device configuration components.
// A device's config should be unique, so the hash should be unique.
//
// These device IDs are not guaranteed to be globally unique, but they should
// be unique to the board they reside on.
func newUID(components ...string) string {
	h := md5.New() // #nosec
	/* #nosec */
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

// registerDevices registers devices with the plugin. Devices are created and
// registered from the unified device configuration, and registered directly
// from dynamic device registration.
func registerDevices() error {

	// devices from dynamic registration
	policy := policies.GetDeviceConfigDynamicPolicy()
	if policy != policies.DeviceConfigDynamicProhibited {
		for _, data := range Config.Plugin.DynamicRegistration.Config {
			devices, err := ctx.dynamicDeviceRegistrar(data)
			if err != nil {
				return err
			}
			log.Debugf("[sdk] adding %d devices from dynamic registration", len(devices))
			updateDeviceMap(devices)
		}
	}

	// devices from config. the config here is the unified device config which
	// is joined from file and from dynamic registration, if set.
	devices, err := makeDevices(Config.Device)
	if err != nil {
		return err
	}
	log.Debugf("[sdk] adding %d devices from config", len(devices))
	updateDeviceMap(devices)

	return nil
}

// logStartupInfo is used to log plugin info at startup. This will log
// the plugin metadata, version info, and registered devices.
func logStartupInfo() {
	// Log plugin metadata
	metainfo.log()
	// Log plugin version info
	version.Log()

	// Log registered devices
	log.Info("Registered Devices:")
	for id, dev := range ctx.devices {
		log.Infof("  %v (%v)", id, dev.Kind)
	}
	log.Info("--------------------------------")
}
