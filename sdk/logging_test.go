package sdk

import (
	"testing"

	"github.com/Sirupsen/logrus"
)

func TestSetLogLevel(t *testing.T) {
	// the default logger level is info, so it should start at info.
	if Logger.Level != logrus.InfoLevel {
		t.Error("Logger did not start at log level INFO")
	}

	SetLogLevel(true)
	if Logger.Level != logrus.DebugLevel {
		t.Error("Failed to set log level to DEBUG")
	}

	SetLogLevel(false)
	if Logger.Level != logrus.InfoLevel {
		t.Error("Failed to set log level back to INFO")
	}
}
