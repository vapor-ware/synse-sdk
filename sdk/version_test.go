package sdk

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestVersionInit tests that the init function initialized things correctly.
func TestVersionInit(t *testing.T) {
	assert.NotNil(t, version)
	assert.Equal(t, "-", version.BuildDate)
	assert.Equal(t, "-", version.GitCommit)
	assert.Equal(t, "-", version.GitTag)
	assert.Equal(t, "-", version.PluginVersion)
	assert.Equal(t, "-", version.GoVersion)
	assert.Equal(t, Version, version.SDKVersion)
	assert.Equal(t, runtime.GOARCH, version.Arch)
	assert.Equal(t, runtime.GOOS, version.OS)
}

// Test_setField tests setting a field value.
func Test_setField(t *testing.T) {
	f := setField("")
	assert.Equal(t, "-", f)

	f = setField("foo")
	assert.Equal(t, "foo", f)
}

// TestBinVersion_Encode tests converting the binVersion to the gRPC VersionInfo.
func TestBinVersion_Encode(t *testing.T) {
	vi := version.Encode()
	assert.NotNil(t, vi)
	assert.Equal(t, "-", vi.PluginVersion)
	assert.Equal(t, "-", vi.GitTag)
	assert.Equal(t, "-", vi.GitCommit)
	assert.Equal(t, "-", vi.BuildDate)
	assert.Equal(t, Version, vi.SdkVersion)
	assert.Equal(t, runtime.GOOS, vi.Os)
	assert.Equal(t, runtime.GOARCH, vi.Arch)
}

// TestBinVersion_Format tests producing a formatted string representation
// of the binVersion.
func TestBinVersion_Format(t *testing.T) {
	// since the values here will change based on when/where this is run,
	// we can only verify that it produces something.
	out := version.Format()
	assert.NotEmpty(t, out)
}

// TestBinVersion_Log tests logging out the binVersion
func TestBinVersion_Log(t *testing.T) {
	version.Log()
}
