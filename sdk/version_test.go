// Synse SDK
// Copyright (c) 2019 Vapor IO
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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

// TestBinVersion_encode tests converting the version to the corresponding gRPC message.
func TestBinVersion_encode(t *testing.T) {
	vi := version.encode()
	assert.NotNil(t, vi)
	assert.Equal(t, "-", vi.PluginVersion)
	assert.Equal(t, "-", vi.GitTag)
	assert.Equal(t, "-", vi.GitCommit)
	assert.Equal(t, "-", vi.BuildDate)
	assert.Equal(t, Version, vi.SdkVersion)
	assert.Equal(t, runtime.GOOS, vi.Os)
	assert.Equal(t, runtime.GOARCH, vi.Arch)
}

// TestBinVersion_format tests producing a formatted string representation
// of the version.
func TestBinVersion_format(t *testing.T) {
	// since the values here will change based on when/where this is run,
	// we can only verify that it produces something.
	out := version.format()
	assert.NotEmpty(t, out)
}

// TestBinVersion_Log tests logging out the binVersion
func TestBinVersion_Log(t *testing.T) {
	version.Log()
}
