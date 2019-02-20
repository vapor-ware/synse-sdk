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

package sdk

// filterDevices returns a list of Devices (a subset of the deviceMap) which
// match the specified filter(s) in the given filter string.
// FIXME: this needs to not use the context to get devices.. use device manager
// FIXME: update to not use kind
func filterDevices(filter string) ([]*Device, error) { // nolint: gocyclo
	//filters := strings.Split(filter, ",")

	var devices []*Device
	//for _, d := range ctx.devices {
	//	devices = append(devices, d)
	//}
	//
	//for _, f := range filters {
	//	pair := strings.Split(f, "=")
	//	if len(pair) != 2 {
	//		return nil, fmt.Errorf("incorrect filter string: %s", filter)
	//	}
	//	k, v := pair[0], pair[1]
	//
	//	var isValid func(d *Device) bool
	//	switch k {
	//	case "kind":
	//		isValid = func(d *Device) bool { return d.Kind == v || v == "*" }
	//	case "type":
	//		isValid = func(d *Device) bool { return d.Type == v || v == "*" }
	//	default:
	//		return nil, fmt.Errorf("unsupported filter key. expect 'kind' but got %s", k)
	//	}
	//
	//	i := 0
	//	for _, d := range devices {
	//		if isValid(d) {
	//			devices[i] = d
	//			i++
	//		}
	//	}
	//	devices = devices[:i]
	//}
	return devices, nil
}
