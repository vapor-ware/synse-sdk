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

package output

// Duration is the output type for a duration reading. This output
// supports the following units:
//   * second (no system)
var Duration = Output{
	Name:      "duration",
	Type:      "duration",
	Precision: 3,
	Units: map[SystemOfMeasure]*Unit{
		NONE: {
			Name:   "second",
			Symbol: "s",
			System: string(NONE),
		},
	},
	Converters: map[SystemOfMeasure]func(value interface{}, to SystemOfMeasure) (interface{}, error){
		// no system(s) to convert between
	},
}
