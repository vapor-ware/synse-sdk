// Synse SDK
// Copyright (c) 2019-2020 Vapor IO
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

package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfigsNotFound(t *testing.T) {
	err := NewConfigsNotFoundError([]string{"foo", "bar"})

	assert.IsType(t, &ConfigsNotFound{}, err)
	assert.Equal(t, 2, len(err.searchPaths))
	assert.Equal(t, "foo", err.searchPaths[0])
	assert.Equal(t, "bar", err.searchPaths[1])
}

func TestConfigsNotFound_Error(t *testing.T) {
	err := NewConfigsNotFoundError([]string{"foo", "bar"})
	out := err.Error()

	assert.Equal(t, "no configuration file(s) found in: [foo bar]", out)
}
