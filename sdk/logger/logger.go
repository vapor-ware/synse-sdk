package logger

import (
	"github.com/Sirupsen/logrus"
)

// Logger is the logger that the SDK uses and that should be used by plugins
// written using the SDK.
var logger = logrus.New()

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
		logger.Level = logrus.DebugLevel
	} else {
		logger.Level = logrus.InfoLevel
	}
}

// Fatal is a wrapper around logger.Fatal.
func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

// Fatalf is a wrapper around logger.Fatalf.
func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

// Error is a wrapper around logger.Error.
func Error(args ...interface{}) {
	logger.Error(args...)
}

// Errorf is a wrapper around logger.Errorf.
func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

// Warn is a wrapper around logger.Warn.
func Warn(args ...interface{}) {
	logger.Warn(args...)
}

// Warnf is a wrapper around logger.Warnf.
func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

// Info is a wrapper around logger.Info.
func Info(args ...interface{}) {
	logger.Info(args...)
}

// Infof is a wrapper around logger.Infof.
func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

// Debug is a wrapper around logger.Debug.
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

// Debugf is a wrapper around logger.Debugf.
func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}
