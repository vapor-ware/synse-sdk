package sdk

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestVersionInit tests that the init function initialized things correctly.
func TestVersionInit(t *testing.T) {
	assert.NotNil(t, Version)
	assert.Equal(t, "-", Version.BuildDate)
	assert.Equal(t, "-", Version.GitCommit)
	assert.Equal(t, "-", Version.GitTag)
	assert.Equal(t, "-", Version.PluginVersion)
	assert.Equal(t, "-", Version.GoVersion)
	assert.Equal(t, SDKVersion, Version.SDKVersion)
	assert.Equal(t, runtime.GOARCH, Version.Arch)
	assert.Equal(t, runtime.GOOS, Version.OS)
}

// Test_setField tests setting a field value.
func Test_setField(t *testing.T) {
	f := setField("")
	assert.Equal(t, "-", f)

	f = setField("foo")
	assert.Equal(t, "foo", f)
}

// TestBinVersion_Encode tests converting the BinVersion to the gRPC VersionInfo.
func TestBinVersion_Encode(t *testing.T) {
	vi := Version.Encode()
	assert.NotNil(t, vi)
	assert.Equal(t, "-", vi.PluginVersion)
	assert.Equal(t, "-", vi.GitTag)
	assert.Equal(t, "-", vi.GitCommit)
	assert.Equal(t, "-", vi.BuildDate)
	assert.Equal(t, SDKVersion, vi.SdkVersion)
	assert.Equal(t, runtime.GOOS, vi.Os)
	assert.Equal(t, runtime.GOARCH, vi.Arch)
}

// TestBinVersion_Format tests producing a formatted string representation
// of the BinVersion.
func TestBinVersion_Format(t *testing.T) {
	// since the values here will change based on when/where this is run,
	// we can only verify that it produces something.
	out := Version.Format()
	assert.NotEmpty(t, out)
}

// TestBinVersion_Log tests logging out the BinVersion
func TestBinVersion_Log(t *testing.T) {
	Version.Log()
}
