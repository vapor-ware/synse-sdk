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

package utils

import (
	"fmt"
	"strconv"
)

// ConvertToFloat64 converts value to a float64 or errors out.
func ConvertToFloat64(value interface{}) (result float64, err error) { // nolint: gocyclo
	switch t := value.(type) {
	case float64:
		result = t
	case float32:
		result = float64(t)
	case int64:
		result = float64(t)
	case int32:
		result = float64(t)
	case int16:
		result = float64(t)
	case int8:
		result = float64(t)
	case int:
		result = float64(t)
	case uint64:
		result = float64(t)
	case uint32:
		result = float64(t)
	case uint16:
		result = float64(t)
	case uint8:
		result = float64(t)
	case uint:
		result = float64(t)
	case string:
		result, err = strconv.ParseFloat(t, 64)
	default:
		err = fmt.Errorf("unable to convert value %v, type %T to float64", value, value)
	}
	return result, err
}
