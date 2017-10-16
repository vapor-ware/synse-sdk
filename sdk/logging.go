package sdk

import (
	"github.com/Sirupsen/logrus"
)

var logger = logrus.New()


func SetLogLevel(debug bool) {
	if debug {
		logger.Level = logrus.DebugLevel
	} else {
		// The highest level we set for the SDK is INFO. This is because there
		// should only be a few instances of INFO logging which could be useful
		// to expose in production. All other production logging should be error
		// logging, which would be captured in this output as well.
		logger.Level = logrus.InfoLevel
	}
}
