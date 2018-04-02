package logger

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestSetLogLevel tests setting the logging level for the SDK logger.
func TestSetLogLevel(t *testing.T) {
	// the default logger level is info, so it should start at info.
	assert.Equal(t, logrus.InfoLevel, logger.Level)

	SetLogLevel(true)
	assert.Equal(t, logrus.DebugLevel, logger.Level)

	SetLogLevel(false)
	assert.Equal(t, logrus.InfoLevel, logger.Level)
}
