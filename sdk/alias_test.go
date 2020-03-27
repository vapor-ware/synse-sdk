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

func TestNewAliasCache(t *testing.T) {
	c := NewAliasCache()
	assert.Empty(t, c.cache)
}

func TestAliasCache_Add_new(t *testing.T) {
	c := AliasCache{
		cache: map[string]*Device{},
	}

	err := c.Add("alias-1", &Device{id: "123"})
	assert.NoError(t, err)
	assert.Len(t, c.cache, 1)
	assert.Contains(t, c.cache, "alias-1")
}

func TestAliasCache_Add_conflict(t *testing.T) {
	c := AliasCache{
		cache: map[string]*Device{
			"alias-1": {id: "123"},
		},
	}

	err := c.Add("alias-1", &Device{id: "456"})
	assert.Error(t, err)
	assert.Len(t, c.cache, 1)
	assert.Contains(t, c.cache, "alias-1")
	assert.Equal(t, c.cache["alias-1"].id, "123")
}

func TestAliasCache_Get_match(t *testing.T) {
	c := AliasCache{
		cache: map[string]*Device{
			"alias-1": {id: "123"},
			"alias-2": {id: "456"},
		},
	}

	device := c.Get("alias-1")
	assert.NotNil(t, device)
	assert.Equal(t, "123", device.id)
}

func TestAliasCache_Get_noMatch(t *testing.T) {
	c := AliasCache{
		cache: map[string]*Device{
			"alias-1": {id: "123"},
			"alias-2": {id: "456"},
		},
	}

	device := c.Get("alias-unknown")
	assert.Nil(t, device)
}
