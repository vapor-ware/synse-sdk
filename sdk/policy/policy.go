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

package policy

// Policy is a string which defines a plugin policy. Policies generally
// dictate whether something like configuration is optional or required.
type Policy string

const (
	// Required designates that an attribute is required.
	Required Policy = "required"

	// Optional designates that an attribute is optional.
	Optional Policy = "optional"
)

// Policies defines all the policies for a plugin.
type Policies struct {
	PluginConfig        Policy
	DeviceConfig        Policy
	DynamicDeviceConfig Policy
}

// NewDefaultPolicies returns an instance of the Policies struct with
// all the default policy values set.
func NewDefaultPolicies() *Policies {
	return &Policies{
		PluginConfig:        Optional,
		DeviceConfig:        Required,
		DynamicDeviceConfig: Optional,
	}
}
