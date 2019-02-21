// synse-sdk
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

// Color is the output type for a color reading. This output has no unit.
// A color reading is generally a string which represents some kind of
// color. This can be the name ("red"), a hex string ("ff0000"), an
// RBG string ("255,0,0"), or anything else.
var Color = Output{
	Name:  "color",
	Type:  "color",
	Units: map[SystemOfMeasure]*Unit{
		// no unit(s)
	},
	Converters: map[SystemOfMeasure]func(value interface{}, to SystemOfMeasure) (interface{}, error){
		// no system(s) to convert between
	},
}
