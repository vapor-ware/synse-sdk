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
