package sdk

import (
	"fmt"

	"runtime"

	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
)

type pluginAction func(p *Plugin) error
type deviceAction func(p *Plugin, d *Device) error

// Plugin represents an instance of a Synse plugin. Along with metadata
// and definable handlers, it contains a gRPC server to handle the plugin
// requests.
type Plugin struct {
	Config   *config.PluginConfig
	server   *Server
	handlers *Handlers
	dm       *DataManager
	v        *VersionInfo

	preRunActions   []pluginAction
	postRunActions  []pluginAction
	devSetupActions map[string][]deviceAction
}

// NewPlugin creates a new Plugin instance
func NewPlugin(handlers *Handlers) *Plugin {
	p := &Plugin{}
	p.handlers = handlers
	p.v = emptyVersionInfo()
	return p
}

// SetVersion sets the VersionInfo for the Plugin.
func (p *Plugin) SetVersion(info VersionInfo) {
	p.v.Merge(&info)
}

// SetConfig manually sets the configuration of the Plugin.
func (p *Plugin) SetConfig(config *config.PluginConfig) error {
	err := config.Validate()
	if err != nil {
		return err
	}
	p.Config = config

	logger.SetLogLevel(p.Config.Debug)
	return nil
}

// Configure reads in the config file and uses it to set the Plugin configuration.
func (p *Plugin) Configure() error {
	cfg, err := config.NewPluginConfig()
	if err != nil {
		return err
	}
	p.Config = cfg
	logger.SetLogLevel(p.Config.Debug)
	return nil
}

// RegisterDevices registers all of the configured devices (via their proto and
// instance config) with the plugin.
func (p *Plugin) RegisterDevices() error {
	return registerDevicesFromConfig(p.handlers.Device, p.Config.AutoEnumerate)
}

// RegisterPreRunActions registers functions with the plugin that will be called
// before the gRPC server and DataManager are started. The functions here can be
// used for plugin-wide setup actions.
func (p *Plugin) RegisterPreRunActions(actions ...pluginAction) {
	if p.preRunActions == nil {
		p.preRunActions = actions
	} else {
		p.preRunActions = append(p.preRunActions, actions...)
	}
}

// RegisterPostRunActions registers functions with the plugin that will be called
// after the gRPC server and DataManager terminate running. The functions here can
// be used for plugin-wide teardown actions.
func (p *Plugin) RegisterPostRunActions(actions ...pluginAction) {
	if p.postRunActions == nil {
		p.postRunActions = actions
	} else {
		p.postRunActions = append(p.postRunActions, actions...)
	}
}

// RegisterDeviceSetupActions registers functions with the plugin that will be
// called on device initialization before it is read from / written to. The
// functions here can be used for device-specific setup actions.
//
// Filter should be the filter to apply to devices. Currently filtering is only
// supported for device type and device model. Filter strings are specified by
// the format "key=value,key=value". "type=temperature,model=ABC123" would only
// match devices whose type was temperature and model was ABC123.
func (p *Plugin) RegisterDeviceSetupActions(filter string, actions ...deviceAction) {
	if p.devSetupActions == nil {
		p.devSetupActions = make(map[string][]deviceAction)
	}
	if _, exists := p.devSetupActions[filter]; exists {
		p.devSetupActions[filter] = append(p.devSetupActions[filter], actions...)
	} else {
		p.devSetupActions[filter] = actions
	}
}

// Run starts the Plugin server which begins listening for gRPC requests.
func (p *Plugin) Run() error {
	err := p.setup()
	if err != nil {
		return err
	}
	p.logInfo()

	// Before we start the DataManager goroutines or the gRPC server, we
	// will execute the preRunActions, if any exist.
	if len(p.preRunActions) > 0 {
		logger.Debug("Executing Pre Run Actions:")
		for _, action := range p.preRunActions {
			logger.Debugf(" * %v", action)
			err := action(p)
			if err != nil {
				return err
			}
		}
	}

	// At this point all state that the plugin will need should be available.
	// With a complete view of the plugin, devices, and configuration, we can
	// now process any device setup actions prior to reading to/writing from
	// the device(s).
	if len(p.devSetupActions) > 0 {
		logger.Debug("Executing Device Setup Actions:")
		for filter, actions := range p.devSetupActions {
			devices, err := filterDevices(filter)
			if err != nil {
				return err
			}
			logger.Debugf("* %v (%v devices match filter %v)", actions, len(devices), filter)
			for _, d := range devices {
				for _, action := range actions {
					err := action(p, d)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	// Start the go routines to poll devices and to update internal state
	// with those readings.
	p.dm.goPollData()
	p.dm.goUpdateData()

	// Start the gRPC server
	return p.server.serve()

	// TODO - figure out how to get post actions working correctly
}

// setup is the pre-run stage where the Plugin handlers and configuration
// are validated and runtime components of the plugin are initialized.
func (p *Plugin) setup() error {
	// validate that handlers are set
	err := validateHandlers(p.handlers)
	if err != nil {
		return err
	}

	// validate that configuration is set
	if p.Config == nil {
		return fmt.Errorf("plugin must be configured before it is run")
	}

	// Setup the transaction cache
	SetupTransactionCache(p.Config.Settings.Transaction.TTL)

	// Register a new Server and DataManager for the Plugin. This should
	// be done prior to running the plugin, as opposed to on initialization
	// of the Plugin struct, because their configuration is configuration
	// dependent. The Plugin should be configured prior to running.
	p.server = NewServer(p)
	p.dm = NewDataManager(p)

	return nil
}

// logInfo logs out the information about the plugin. This is called just before the
// plugin begins running all of its components.
func (p *Plugin) logInfo() {
	logger.Info("Plugin Info:")
	logger.Infof(" Name:        %s", p.Config.Name)
	logger.Infof(" Version:     %s", p.v.VersionString)
	logger.Infof(" SDK Version: %s", SDKVersion)
	logger.Infof(" Git Commit:  %s", p.v.GitCommit)
	logger.Infof(" Git Tag:     %s", p.v.GitTag)
	logger.Infof(" Go Version:  %s", p.v.GoVersion)
	logger.Infof(" Build Date:  %s", p.v.BuildDate)
	logger.Infof(" OS:          %s", runtime.GOOS)
	logger.Infof(" Arch:        %s", runtime.GOARCH)
	logger.Debug("Plugin Config:")
	logger.Debugf(" %#v", p.Config)
	logger.Info("Registered Devices:")
	for id, dev := range deviceMap {
		logger.Infof(" %v (%v)", id, dev.Model())
	}
	logger.Info("--------------------------------")
}
