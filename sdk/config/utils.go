package config

import "fmt"

// toSliceStringMapI casts an interface to a []map[string]interface{} type. This is used
// to convert the Viper []interface{} type read in for the 'auto_enumerate' config field
// to the appropriate type in the PluginConfig struct.
func toSliceStringMapI(i interface{}) ([]map[string]interface{}, error) {
	var s = []map[string]interface{}{}

	switch t := i.(type) {
	case []interface{}:
		for _, l := range t {
			m := map[string]interface{}{}
			for k, v := range l.(map[interface{}]interface{}) {
				m[k.(string)] = v
			}
			s = append(s, m)
		}
	case []map[string]interface{}:
		s = t
	default:
		return s, fmt.Errorf("unable to cast %#v of type %T to []map[string]interface{}", i, i)
	}
	return s, nil
}

// toStringMapI casts an interface to a map[string]interface{} type. This is used
// to convert the Viper map[interface{}]interface{} type read in for the 'context'
// config field to the appropriate type in the PluginConfig struct.
func toStringMapI(i interface{}) (map[string]interface{}, error) {
	var m = map[string]interface{}{}

	switch t := i.(type) {
	case map[interface{}]interface{}:
		for k, v := range t {
			m[k.(string)] = v
		}
	case map[string]interface{}:
		m = t
	default:
		return m, fmt.Errorf("unable to cast %#v of type %T to map[string]interface{}", i, i)
	}
	return m, nil
}
