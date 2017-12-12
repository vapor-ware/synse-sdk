package sdk

import (
	"os"
)

const (
	// FIXME - this is probably not the right place for this const to be defined.
	defaultConfigFile = "/etc/synse/plugin/config.yml"
)

// Plugin represents an instance of a Synse plugin. Along with metadata
// and definable handlers, it contains a gRPC server to handle the plugin
// requests.
type Plugin struct {
	server   *Server
	handlers *Handlers
	dm       *DataManager
}

// NewPlugin creates a new Plugin instance
func NewPlugin(handlers *Handlers) *Plugin {
	p := &Plugin{}
	p.handlers = handlers
	return p
}

// SetConfig sets the configuration of the Plugin.
func (p *Plugin) SetConfig(config *PluginConfig) error {
	return configurePlugin(config)
}

// ConfigureFromFile sets the Plugin configuration from the specified YAML file.
func (p *Plugin) ConfigureFromFile(path string) error {
	config := PluginConfig{}
	err := config.FromFile(path)
	if err != nil {
		return err
	}
	return configurePlugin(&config)
}

// FIXME - this can probably be tidied up somehow / the logic moved to the 'config' file
// Configure reads in the specified config file and uses its contents
// to configure the Plugin.
func (p *Plugin) Configure() error {
	config := PluginConfig{}

	configFile := os.Getenv("PLUGIN_CONFIG")
	if configFile == "" {
		configFile = defaultConfigFile
	}
	err := config.FromFile(configFile)
	if err != nil {
		return err
	}
	return configurePlugin(&config)
}

func (p *Plugin) RegisterDevices() error {
	return registerDevicesFromConfig(p.handlers.Device)
}

// Run starts the Plugin server which begins listening for gRPC requests.
func (p *Plugin) Run() error {
	err := p.setup()
	if err != nil {
		return err
	}

	// Start the go routines to poll devices and to update internal state
	// with those readings.
	p.dm.goPollData()
	p.dm.goUpdateData()

	// Start the gRPC server
	err = p.server.serve()
	if err != nil {
		return err
	}
	return nil
}

// setup is the pre-run stage where the Plugin handlers and configuration
// are validated and runtime components of the plugin are initialized.
func (p *Plugin) setup() error {
	// validate that handlers are set

	// validate that configuration is set

	// Register a new Server and DataManager for the Plugin. This should
	// be done prior to running the plugin, as opposed to on initialization
	// of the Plugin struct, because their configuration is configuration
	// dependent. The Plugin should be configured prior to running.
	p.server = NewServer(p)
	p.dm = NewDataManager()

	return nil
}
