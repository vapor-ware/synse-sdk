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

// Frequency is the output type for a frequency reading. This output
// supports the following units:
//   * Hertz (no system)
var Frequency = Output{
	Name:      "frequency",
	Type:      "frequency",
	Precision: 2,
	Units: map[SystemOfMeasure]*Unit{
		NONE: {
			Name:   "hertz",
			Symbol: "Hz",
			System: string(NONE),
		},
	},
	Converters: map[SystemOfMeasure]func(value interface{}, to SystemOfMeasure) (interface{}, error){
		// no system(s) to convert between
	},
}
