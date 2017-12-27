package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// extensions supported for device configuration files
var supportedExts = []string{".yml", ".yaml"}

// isValidConfig checks if the given FileInfo corresponds to a file that could be
// a valid configuration file. It checks that it is actually a file (not a Dir)
// and checks that its extension matches the supported extensions.
func isValidConfig(f os.FileInfo) bool {
	if f.IsDir() {
		return false
	}

	ok := false
	for _, ext := range supportedExts {
		if filepath.Ext(f.Name()) == ext {
			ok = true
		}
	}
	return ok
}

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
