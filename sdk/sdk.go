package sdk

const (

	// ConfigDir is the directory which contains the device configurations.
	configDir = "config"
)

// NewPlugin creates a new SDK PluginServer instance. This is the preferred way
// of creating a new PluginServer. When called, this will read in configuration
// files present in the plugin config directory and use those configurations to
// set up a new instance of the PluginServer.
func NewPlugin(config *PluginConfig, pluginHandler PluginHandler, deviceHandler DeviceHandler) (*PluginServer, error) {

	err := configurePlugin(config)
	if err != nil {
		return nil, err
	}
	SetLogLevel(Config.Debug)

	readChan := make(chan *ReadResource, Config.Settings.Read.BufferSize)
	writeChan := make(chan *WriteResource, Config.Settings.Write.BufferSize)

	readManager := ReadingManager{
		channel: readChan,
		values:  make(map[string][]*Reading),
	}

	writeManager := WritingManager{
		channel: writeChan,
	}

	rwloop := RWLoop{
		handler:        pluginHandler,
		readingManager: readManager,
		writingManager: writeManager,
	}

	Logger.Debugf("Config: %+v", Config)

	s := &PluginServer{
		name:           Config.Name,
		rwLoop:         rwloop,
		readingManager: readManager,
		writingManager: writeManager,
	}

	Logger.Info("[plugin] new plugin instance created")

	// read in the device prototype and instance configurations from the plugin's
	// config files and generate devices for each configured device.
	err = s.configureDevices(deviceHandler)
	if err != nil {
		return nil, err
	}

	Logger.Infof("[plugin] registered %v devices", len(s.pluginDevices))
	return s, nil
}
