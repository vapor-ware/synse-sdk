package sdk

import (
	"fmt"
)


const (
	CONFIG_DIR = "config"
)


// create a new SDK Plugin instance. This is the preferred way of creating a
// new PluginServer object. When called, this will read in configurations
// present in the plugin config directory for prototype and instance configs.
func NewPlugin(name string, pluginHandler PluginHandler, deviceHandler DeviceHandler) *PluginServer {

	// FIXME: configurable buffer size?
	readChan := make(chan ReadResource, 100)
	writeChan := make(chan WriteResource, 100)

	readManager := ReadingManager{
		channel: readChan,
		values: make(map[string][]Reading),
	}

	writeManager := WritingManager{
		channel: writeChan,
		values: make(map[string]string),
	}

	rwloop := RWLoop{
		handler: pluginHandler,
		readingManager: readManager,
		writingManager: writeManager,
	}


	s := &PluginServer{
		name: name,
		rwLoop: rwloop,
		readingManager: readManager,
		writingManager: writeManager,
	}

	fmt.Printf("[plugin] new plugin instance created\n")

	// read in the device prototype and instance configurations from the plugin's
	// config files and generate devices for each configured device.
	s.configureDevices(deviceHandler)
	fmt.Printf("[plugin] registered %v devices\n", len(s.pluginDevices))

	return s
}

