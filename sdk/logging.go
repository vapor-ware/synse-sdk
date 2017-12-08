package sdk

import (
	"github.com/Sirupsen/logrus"
)

// Logger is the logger that the SDK uses and that should be used by plugins
// written using the SDK.
var Logger = logrus.New()

// SetLogLevel sets the level of `Logger` to either debug or info based on
// the debug boolean flag passed to it.
//
// While more levels could be supported, we really only care about logging
// for development environments and logging for production. Production
// logging happens at the info level as opposed to error, since there are
// informational messages that can be helpful to surface in production,
// not just error or warning messages.
func SetLogLevel(debug bool) {
	if debug {
		Logger.Level = logrus.DebugLevel
	} else {
		Logger.Level = logrus.InfoLevel
	}
}
