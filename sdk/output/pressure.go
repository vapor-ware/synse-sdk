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

import (
	"fmt"

	"github.com/vapor-ware/synse-sdk/sdk/utils"
)

// Pressure is the output type for a pressure reading. This output
// supports the following units:
//   * Pascal (metric)
//   * Pound-force per square inch (imperial)
var Pressure = Output{
	Name:      "pressure",
	Type:      "pressure",
	Precision: 3,
	Units: map[SystemOfMeasure]*Unit{
		METRIC: {
			Name:   "pascal",
			Symbol: "Pa",
			System: string(METRIC),
		},
		IMPERIAL: {
			Name:   "pound-force per square inch",
			Symbol: "psi",
			System: string(IMPERIAL),
		},
	},
	Converters: map[SystemOfMeasure]func(value interface{}, to SystemOfMeasure) (interface{}, error){
		METRIC:   fromPascal,
		IMPERIAL: fromPoundForcePerSquareInch,
	},
}

func fromPascal(value interface{}, to SystemOfMeasure) (interface{}, error) {
	switch to {
	case METRIC:
		return value, nil

	case IMPERIAL:
		asFloat, err := utils.ConvertToFloat64(value)
		if err != nil {
			return nil, err
		}
		converted := float64(asFloat / 6894.757) // approximate result
		return converted, nil

	default:
		// todo; common error
		return nil, fmt.Errorf("invalid system")
	}
}

func fromPoundForcePerSquareInch(value interface{}, to SystemOfMeasure) (interface{}, error) {
	switch to {
	case METRIC:
		asFloat, err := utils.ConvertToFloat64(value)
		if err != nil {
			return nil, err
		}
		converted := float64(asFloat * 6894.757) // approximate result
		return converted, nil

	case IMPERIAL:
		return value, nil

	default:
		// todo; common error
		return nil, fmt.Errorf("invalid system")
	}
}
