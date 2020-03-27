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

package funcs

import "github.com/vapor-ware/synse-sdk/sdk/utils"

// GetBuiltins returns all of the built-in Funcs supplied by the SDK.
func GetBuiltins() []*Func {
	return []*Func{
		&FtoC,
	}
}

// FtoC is a Func which converts a value from degrees Fahrenheit to
// degrees Celsius.
var FtoC = Func{
	Name: "FtoC",
	Fn: func(value interface{}) (interface{}, error) {
		f, err := utils.ConvertToFloat64(value)
		if err != nil {
			return nil, err
		}
		c := float64((f - 32.0) * 5.0 / 9.0)
		return c, nil
	},
}
