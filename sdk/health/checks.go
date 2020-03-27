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

package health

// CheckType defines the type of a health Check.
type CheckType string

// Check is an interface which all health checks should implement.
type Check interface {

	// GetName gets the human-readable name of the health check.
	GetName() string

	// GetType gets the type of the health check. Currently, the supported
	// types are: periodic.
	GetType() CheckType

	// Status gets the latest status of the health check. The status tells
	// whether the health check passed or not at a given time.
	Status() *Status

	// Update the state of the health check. The behavior/timing of how
	// Update is called is based on the type of the check. For example,
	// periodic checks run the Update function on a timed interval.
	Update()

	// Run starts the health check. This is where the update call behavior is
	// defined.
	Run()
}
