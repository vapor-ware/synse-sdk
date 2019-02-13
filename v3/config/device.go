package config

type Device struct {
	Version int
	Devices []DeviceProto
}

type DeviceProto struct {
	Type      string
	Metadata  map[string]string
	Tags      []string
	Data      map[string]interface{}
	Handler   string
	Instances []DeviceInstances
	System    string
}

type DeviceInstances struct {
	Info      string
	Tags      []string
	Data      map[string]interface{}
	Output    string
	SortIndex int32
	Handler   string

	ScalingFactor string
	System        string

	DisableInheritance bool
}
