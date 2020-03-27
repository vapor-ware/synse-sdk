// Synse SDK
// Copyright (c) 2017-2020 Vapor IO
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
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSetPluginInfo tests setting the global plugin meta-information.
func TestSetPluginInfo(t *testing.T) {
	// Make sure it is empty to begin with
	assert.IsType(t, PluginMetadata{}, metadata)
	assert.Equal(t, "", metadata.Name)
	assert.Equal(t, "", metadata.Maintainer)
	assert.Equal(t, "", metadata.Description)
	assert.Equal(t, "", metadata.VCS)

	// Set the metadata
	SetPluginInfo("name", "maintainer", "desc", "vcs")

	// Check that it has changed
	assert.IsType(t, PluginMetadata{}, metadata)
	assert.Equal(t, "name", metadata.Name)
	assert.Equal(t, "maintainer", metadata.Maintainer)
	assert.Equal(t, "desc", metadata.Description)
	assert.Equal(t, "vcs", metadata.VCS)
}

// TestPluginMetadata_log tests logging out the metadata.
func TestPluginMetadata_log(t *testing.T) {
	metadata.log()
}

// TestPluginMetadata_Tag tests making metadata tags.
func TestPluginMetadata_Tag(t *testing.T) {
	var cases = []struct {
		name       string
		maintainer string
		expected   string
	}{
		{
			name:       "test",
			maintainer: "vapor io",
			expected:   "vapor-io/test",
		},
		{
			name:       "Test",
			maintainer: "vaporio",
			expected:   "vaporio/test",
		},
		{
			name:       "Simple Plugin",
			maintainer: "Vapor I-0",
			expected:   "vapor-i_0/simple-plugin",
		},
		{
			name:       "Simple Modbus-over-IP",
			maintainer: "Vapor IO",
			expected:   "vapor-io/simple-modbus_over_ip",
		},
		{
			name:       "99 bottles of beer",
			maintainer: "The Wall",
			expected:   "the-wall/99-bottles-of-beer",
		},
	}

	for i, c := range cases {
		meta := PluginMetadata{
			Name:       c.name,
			Maintainer: c.maintainer,
		}

		assert.Equal(t, c.expected, meta.Tag(), "case: %d", i)
	}
}

func TestPluginMetadata_encode(t *testing.T) {
	m := PluginMetadata{
		Name:        "test",
		Maintainer:  "vaporio",
		Description: "test metadata",
	}

	encoded := m.encode()
	assert.Equal(t, "test", encoded.Name)
	assert.Equal(t, "vaporio", encoded.Maintainer)
	assert.Equal(t, "test metadata", encoded.Description)
	assert.Equal(t, "vaporio/test", encoded.Tag)
	assert.Equal(t, "", encoded.Vcs)
}

func TestPluginMetadata_format(t *testing.T) {
	m := PluginMetadata{
		Name:        "test",
		Maintainer:  "vaporio",
		Description: "test metadata",
	}

	out := m.format()
	assert.Equal(t, `Plugin Info:
  Name:        test
  Maintainer:  vaporio
  VCS:         
  Description: test metadata`, out)
}
