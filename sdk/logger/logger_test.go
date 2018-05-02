package logger

import (
	"bytes"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// setupLogger is a test utility function to set up the logger
// to write out to a bytes buffer for tests.
func setupLogger(level logrus.Level) *bytes.Buffer {
	// Set the logger output to a bytes buffer
	var buffer bytes.Buffer
	logger.Out = &buffer

	// Set the level
	logger.Level = level

	// Set a custom formatter for testing - this is to disable the
	// timestamp in the log to make it easier to test the output value.
	logger.Formatter = &logrus.TextFormatter{
		DisableTimestamp: true,
	}

	return &buffer
}

// TestSetLogLevel tests setting the logging level for the SDK logger.
func TestSetLogLevel(t *testing.T) {
	// the default logger level is info, so it should start at info.
	assert.Equal(t, logrus.InfoLevel, logger.Level)

	SetLogLevel(true)
	assert.Equal(t, logrus.DebugLevel, logger.Level)

	SetLogLevel(false)
	assert.Equal(t, logrus.InfoLevel, logger.Level)
}

func TestDebug(t *testing.T) {
	buffer := setupLogger(logrus.DebugLevel)

	Debug("test")
	assert.Equal(t, "level=debug msg=test\n", buffer.String())
}

func TestDebug_BadLevel(t *testing.T) {
	buffer := setupLogger(logrus.FatalLevel)

	Debug("test")
	assert.Equal(t, "", buffer.String())
}

func TestDebugf(t *testing.T) {
	buffer := setupLogger(logrus.DebugLevel)

	Debugf("test %d", 1)
	assert.Equal(t, "level=debug msg=\"test 1\"\n", buffer.String())
}

func TestDebugf_BadLevel(t *testing.T) {
	buffer := setupLogger(logrus.FatalLevel)

	Debugf("test %d", 1)
	assert.Equal(t, "", buffer.String())
}

func TestInfo(t *testing.T) {
	buffer := setupLogger(logrus.InfoLevel)

	Info("test")
	assert.Equal(t, "level=info msg=test\n", buffer.String())
}

func TestInfo_BadLevel(t *testing.T) {
	buffer := setupLogger(logrus.FatalLevel)

	Info("test")
	assert.Equal(t, "", buffer.String())
}

func TestInfof(t *testing.T) {
	buffer := setupLogger(logrus.InfoLevel)

	Infof("test %d", 1)
	assert.Equal(t, "level=info msg=\"test 1\"\n", buffer.String())
}

func TestInfof_BadLevel(t *testing.T) {
	buffer := setupLogger(logrus.FatalLevel)

	Infof("test %d", 1)
	assert.Equal(t, "", buffer.String())
}

func TestInfoMultiline_EmptyLine(t *testing.T) {
	buffer := setupLogger(logrus.InfoLevel)

	InfoMultiline("")
	assert.Equal(t, "level=info msg=\"Line 0000: \"\n", buffer.String())
}

func TestInfoMultiline_SingleLine(t *testing.T) {
	buffer := setupLogger(logrus.InfoLevel)

	InfoMultiline("test")
	assert.Equal(t, "level=info msg=\"Line 0000: test\"\n", buffer.String())
}

func TestInfoMultiline_MultiLine(t *testing.T) {
	buffer := setupLogger(logrus.InfoLevel)

	InfoMultiline("one\ntwo\nthree")
	assert.Equal(t, "level=info msg=\"Line 0000: one\"\nlevel=info msg=\"Line 0001: two\"\nlevel=info msg=\"Line 0002: three\"\n", buffer.String())
}

func TestInfoMultiline_BadLevel(t *testing.T) {
	buffer := setupLogger(logrus.FatalLevel)

	InfoMultiline("test")
	assert.Equal(t, "", buffer.String())
}

func TestWarn(t *testing.T) {
	buffer := setupLogger(logrus.WarnLevel)

	Warn("test")
	assert.Equal(t, "level=warning msg=test\n", buffer.String())
}

func TestWarn_BadLevel(t *testing.T) {
	buffer := setupLogger(logrus.FatalLevel)

	Warn("test")
	assert.Equal(t, "", buffer.String())
}

func TestWarnf(t *testing.T) {
	buffer := setupLogger(logrus.WarnLevel)

	Warnf("test %d", 1)
	assert.Equal(t, "level=warning msg=\"test 1\"\n", buffer.String())
}

func TestWarnf_BadLevel(t *testing.T) {
	buffer := setupLogger(logrus.FatalLevel)

	Warnf("test %d", 1)
	assert.Equal(t, "", buffer.String())
}

func TestError(t *testing.T) {
	buffer := setupLogger(logrus.ErrorLevel)

	Error("test")
	assert.Equal(t, "level=error msg=test\n", buffer.String())
}

func TestError_BadLevel(t *testing.T) {
	buffer := setupLogger(logrus.FatalLevel)

	Error("test")
	assert.Equal(t, "", buffer.String())
}

func TestErrorf(t *testing.T) {
	buffer := setupLogger(logrus.ErrorLevel)

	Errorf("test %d", 1)
	assert.Equal(t, "level=error msg=\"test 1\"\n", buffer.String())
}

func TestErrorf_BadLevel(t *testing.T) {
	buffer := setupLogger(logrus.FatalLevel)

	Errorf("test %d", 1)
	assert.Equal(t, "", buffer.String())
}
