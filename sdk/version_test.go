package sdk

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetVersion gets the version and verifies that all fields are
// set correctly. In this case, we do not have all variables set, so
// we expect some fields to contain the default "empty" string.
func TestGetVersion(t *testing.T) {
	version := GetVersion()

	// These fields are not set via global vars, so they should be empty
	assert.Equal(t, "-", version.BuildDate)
	assert.Equal(t, "-", version.GitCommit)
	assert.Equal(t, "-", version.GitTag)
	assert.Equal(t, "-", version.GoVersion)
	assert.Equal(t, "-", version.PluginVersion)

	// These should be set.
	assert.Equal(t, runtime.GOOS, version.OS)
	assert.Equal(t, runtime.GOARCH, version.Arch)
	assert.Equal(t, SDKVersion, version.SDKVersion)
}

// TestGetVersion2 gets the version info when there are some variable set.
func TestGetVersion2(t *testing.T) {
	GitCommit = "123"
	GitTag = "456"
	PluginVersion = "1.2.3"

	version := GetVersion()

	// These fields are not set via global vars, so they should be empty
	assert.Equal(t, "-", version.BuildDate)
	assert.Equal(t, "-", version.GoVersion)

	// These should be set.
	assert.Equal(t, GitCommit, version.GitCommit)
	assert.Equal(t, GitTag, version.GitTag)
	assert.Equal(t, PluginVersion, version.PluginVersion)
	assert.Equal(t, runtime.GOOS, version.OS)
	assert.Equal(t, runtime.GOARCH, version.Arch)
	assert.Equal(t, SDKVersion, version.SDKVersion)
}

// TestBinVersion_Log tests logging out the version info.
func TestBinVersion_Log(t *testing.T) {
	// TODO - for now, we just log to make sure nothing goes wrong. We should
	// really get the logger output and compare to it.

	version := GetVersion()
	version.Log()
}
