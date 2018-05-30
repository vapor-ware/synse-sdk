package device

import (
	"crypto/md5"
	"fmt"
	"io"
	"strings"
)

// FIXME: these function names should probably be changed to something more descriptive

// makeIDString makes a compound string out of the given rack, board, and
// device identifier strings. This string should be a globally unique identifier
// for a given device.
func makeIDString(rack, board, device string) string {
	return strings.Join([]string{rack, board, device}, "-")
}

// newUID creates a new unique identifier for a device. The device id is
// deterministic because it is created as a hash of various components that
// make up the device's configuration. By definition, each device will have
// a (slightly) different configuration (otherwise they would just be the same
// devices).
//
// These device IDs are not guaranteed to be globally unique, but they should
// be unique to the board they reside on.
func newUID(components ...string) string {
	h := md5.New() // nolint: gas
	for _, comp := range components {
		io.WriteString(h, comp) // nolint: errcheck
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
