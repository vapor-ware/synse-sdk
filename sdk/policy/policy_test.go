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

package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultPolicies(t *testing.T) {
	p := NewDefaultPolicies()

	assert.Equal(t, Optional, p.PluginConfig)
	assert.Equal(t, Required, p.DeviceConfig)
	assert.Equal(t, Optional, p.DynamicDeviceConfig)
}
