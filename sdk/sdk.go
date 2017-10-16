package sdk



const (
	CONFIG_DIR = "config"
)


// create a new SDK Plugin instance. This is the preferred way of creating a
// new PluginServer object. When called, this will read in configurations
// present in the plugin config directory for prototype and instance configs.
func NewPlugin(config PluginConfig, pluginHandler PluginHandler, deviceHandler DeviceHandler) (*PluginServer, error) {

	err := ConfigurePlugin(config); if err != nil {
		return nil, err
	}
	SetLogLevel(Config.Debug)

	readChan := make(chan ReadResource, Config.ReadBufferSize)
	writeChan := make(chan WriteResource, Config.WriteBufferSize)

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

	logger.Debugf("Config: %+v", Config)

	s := &PluginServer{
		name: Config.Name,
		rwLoop: rwloop,
		readingManager: readManager,
		writingManager: writeManager,
	}

	logger.Info("[plugin] new plugin instance created")

	// read in the device prototype and instance configurations from the plugin's
	// config files and generate devices for each configured device.
	s.configureDevices(deviceHandler)
	logger.Infof("[plugin] registered %v devices", len(s.pluginDevices))

	return s, nil
}

