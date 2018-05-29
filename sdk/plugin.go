package sdk

import (
	"encoding/json"
	"fmt"

	"flag"
	"os"

	"github.com/vapor-ware/synse-sdk/sdk/cfg"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
	"github.com/vapor-ware/synse-sdk/sdk/policies"
)

type pluginAction func(p *Plugin) error
type deviceAction func(p *Plugin, d *Device) error

// NPlugin is the New Plugin. This will replace Plugin once v1.0 work is completed.
type NPlugin struct {
	policies []policies.ConfigPolicy
}

// Run starts the Plugin.
//
// Before the gRPC server is started, and before the read and write goroutines
// are started, Plugin setup and validation will happen. If successful, pre-run
// actions are executed, and device setup actions are executed, if defined.
func (plugin *NPlugin) Run() {

	// ** "Config" steps **

	// Check for command line flags. If any flags are set that require an
	// action, that action will be resolved here.
	plugin.resolveFlags()

	// Check for configuration policies. If no policy was set by the plugin,
	// this will fall back on the default policies.
	plugin.checkPolicies()

	// Read in all configs and verify that they are correct.
	plugin.processConfig()

	// ** "Making" steps **

	// FIXME: at this point, we will need to resolve dynamic registration stuff.
	//   - if dynamic registration makes Devices, just add to the global map.
	//   - if dynamic registration makes Config artifacts, add to the unified config.
	// (or maybe this happens in th processConfig step, before unification?)

	// Initialize Device instances for each of the devices configured with
	// the plugin.
	plugin.makeDevices()

	// makeTransactionCache
	// makeDataManager
	// makeServer

	// ** "Action" steps **

	// preRunActions
	// deviceSetupActions

	// If the --dry-run flag is set, we will end here. The gRPC server and
	// data manager do not get started up in the dry run.
	if !flagDryRun {

		// "Starting" steps **

		// startDataManager
		// startServer

	}

	// profit

}

// FIXME: its not clear that these private functions need to hang off the Plugin struct..

// resolveFlags parses flags passed to the plugin.
//
// Not all flags will result in an action, so not all flags are checked
// here. Only the flags that cause immediate actions will be processed here.
// Only SDK supported flags are parsed here. If a plugin specifies additional
// flags, they should be resolved in their own pre-run action.
func (plugin *NPlugin) resolveFlags() {
	flag.Parse()

	// --help is already provided by the flag package, so we don't have
	// to handle that here.

	// Print out the version info for the plugin.
	if flagVersion {
		versionInfo := GetVersion()
		fmt.Print(versionInfo.Format())
		os.Exit(0)
	}
}

// checkPolicies checks for policies registered with the plugin. If no policies
// were set, the defaults will be used.
func (plugin *NPlugin) checkPolicies() {
	// Verify that the policies set for the plugin do not break any of the
	// constraints set on the policies.
	err := policies.CheckConstraints(plugin.policies)
	if err.Err() != nil {
		logger.Error("config policies set for the plugin are invalid")
		logger.Fatal(err)
	}

	// If we passed constraint checking, we can use the given policies
	policies.Set(plugin.policies)
}

// processConfig handles plugin configuration in a number of steps. The behavior
// of config handling is dependent on the config policy that is set. If no config
// policies are set, the plugin will terminate in error.
//
// There are four major steps to processing plugin configuration: reading in the
// config, validating the config scheme, config unification, and verifying the
// config data is correct. These steps should happen for all config types.
func (plugin *NPlugin) processConfig() {
	// 1. Read in the configs. We have a few types of configs to read in.
	//  a. Plugin Config
	pluginCtx, err := cfg.GetPluginConfigFromFile()
	if err != nil {
		// what do we do when we error out here?
	}

	//  b. Device Config
	deviceCtxs, err := cfg.GetDeviceConfigsFromFile()
	if err != nil {
		// what do we do when we error out here?
	}

	//  c. Type Config
	// TODO

	//  d. ... Other?
	// TODO?

	// 2. Validate Config Schemes
	multiErr := cfg.Validator.Validate(pluginCtx, pluginCtx.Source)
	if multiErr.HasErrors() {
		// what to do here?
	}

	for _, ctx := range deviceCtxs {
		multiErr = cfg.Validator.Validate(ctx, ctx.Source)
		if multiErr.HasErrors() {
			// what to do here?
		}
	}

	// 3. Unify Configs
	unifiedCtx, err := cfg.UnifyDeviceConfigs(deviceCtxs)
	if err != nil {
		// what to do here?
	}

	// 4. Verify
	if !unifiedCtx.IsDeviceConfig() {
		// ERROR
	}
	unifiedCfg := unifiedCtx.Config.(*cfg.DeviceConfig)
	multiErr = cfg.VerifyConfigs(unifiedCfg)
	if multiErr.HasErrors() {
		// what to do here?
	}

	// TODO: once everything is done, what do we do with all the data?
}

// makeDevices converts the plugin configuration into the Device instances that
// represent the physical/virtual devices that the plugin will manage.
func (plugin *NPlugin) makeDevices() {
	// TODO: implement
}

// Plugin represents an instance of a Synse plugin. Along with metadata
// and definable handlers, it contains a gRPC server to handle the plugin
// requests.
type Plugin struct {
	Config      *config.PluginConfig // See config.PluginConfig for comments.
	server      *server              // InternalApiServer for fulfilling gRPC requests.
	handlers    *Handlers            // See sdk.handlers.go for comments.
	dataManager *dataManager         // Manages device reads and writes. Accesses cached read data.

	deviceHandlers     []*DeviceHandler          // Plugin-specific read and write functions for the devices supported by the plugin.
	preRunActions      []pluginAction            // Array of pluginAction to execute before the main plugin loop.
	postRunActions     []pluginAction            // Array of pluginAction to execute after the main plugin loop.
	deviceSetupActions map[string][]deviceAction // See comments for RegisterDeviceSetupActions.
}

// NewPlugin creates a new Plugin instance. This is the preferred way of
// initializing a new Plugin instance.
//
// The handlers parameter is required and must not be nil. The pluginConfig
// parameter may be nil; if it is, the SDK will attempt to load the
// configuration from file in: /etc/synse/plugin, $HOME/.synse/plugin,
// and $PWD.
func NewPlugin(handlers *Handlers, pluginConfig *config.PluginConfig) (*Plugin, error) {
	logger.SetLogLevel(true)

	// Parameter checks.
	if handlers == nil {
		return nil, invalidArgumentErr("handlers parameter must not be nil")
	}
	// PluginConfig may be nil.

	// Create the Plugin.
	p := &Plugin{}
	p.handlers = handlers

	// If a configuration is passed in, use it.
	// If not, default to finding the config in files.
	if pluginConfig != nil {
		logger.Infof("Using plugin config from parameter: %v", pluginConfig)
		p.Config = pluginConfig
	} else {
		logger.Info("Loading plugin config from file")
		cnfg, err := config.NewPluginConfig()
		if err != nil {
			logger.Errorf("Failed to load plugin config from file: %v", err)
			return nil, err
		}
		p.Config = cnfg
	}
	// Set logging level from the config now that we have a config.
	logger.SetLogLevel(p.Config.Debug)

	return p, nil
}

// RegisterHandlers registers device handlers for the plugin.
func (p *Plugin) RegisterHandlers(handlers *Handlers) {
	p.handlers = handlers
}

// RegisterDeviceIdentifier sets the given identifier function as the DeviceIdentifier
// handler for the plugin. This function helps generate the device UID by letting the
// SDK know which pieces of a Device instance's configuration are unique to that device.
func (p *Plugin) RegisterDeviceIdentifier(identifier DeviceIdentifier) {
	p.handlers.DeviceIdentifier = identifier
}

// RegisterDeviceEnumerator sets the given enumerator function as the DeviceEnumerator
// handler for the plugin.
func (p *Plugin) RegisterDeviceEnumerator(enumerator DeviceEnumerator) {
	p.handlers.DeviceEnumerator = enumerator
}

// RegisterDeviceHandlers adds DeviceHandlers to the Plugin.
//
// These DeviceHandlers are then matched with the Device instances
// by their type/model and provide the read/write functionality for the
// Devices. If a DeviceHandler for a Device is not registered here, the
// Device will not be usable by the plugin.
func (p *Plugin) RegisterDeviceHandlers(handlers ...*DeviceHandler) {
	if p.deviceHandlers == nil {
		p.deviceHandlers = handlers
	} else {
		p.deviceHandlers = append(p.deviceHandlers, handlers...)
	}
}

// SetConfig manually sets the configuration of the Plugin. This is
// generally not the recommended way to configure a Plugin.
func (p *Plugin) SetConfig(config *config.PluginConfig) error {
	err := config.Validate()
	if err != nil {
		logger.Errorf("Failed plugin config validation: %v", err)
		return err
	}
	p.Config = config

	logger.SetLogLevel(p.Config.Debug)
	return nil
}

// registerDevices registers all of the configured devices (via their proto and
// instance config) with the plugin.
func (p *Plugin) registerDevices() error {
	var devices []*config.DeviceConfig

	cfgDevices, err := devicesFromConfig()
	if err != nil {
		logger.Errorf("Failed to register devices from files: %v", err)
		return err
	}
	devices = append(devices, cfgDevices...)

	enumDevices, err := devicesFromAutoEnum(p)
	if err != nil {
		logger.Errorf("Failed to register devices from auto-enum: %v", err)
		return err
	}
	devices = append(devices, enumDevices...)

	return registerDevices(p, devices)
}

// RegisterPreRunActions registers functions with the plugin that will be called
// before the gRPC server and dataManager are started. The functions here can be
// used for plugin-wide setup actions.
func (p *Plugin) RegisterPreRunActions(actions ...pluginAction) {
	if p.preRunActions == nil {
		p.preRunActions = actions
	} else {
		p.preRunActions = append(p.preRunActions, actions...)
	}
}

// RegisterPostRunActions registers functions with the plugin that will be called
// after the gRPC server and dataManager terminate running. The functions here can
// be used for plugin-wide teardown actions.
//
// NOTE: While post run actions can be defined for a Plugin, they are currently
// not executed. See: https://github.com/vapor-ware/synse-sdk/issues/85
func (p *Plugin) RegisterPostRunActions(actions ...pluginAction) {
	if p.postRunActions == nil {
		p.postRunActions = actions
	} else {
		p.postRunActions = append(p.postRunActions, actions...)
	}
}

// RegisterDeviceSetupActions registers functions with the plugin that will be
// called on device initialization before it is ever read from / written to. The
// functions here can be used for device-specific setup actions.
//
// The filter parameter should be the filter to apply to devices. Currently
// filtering is only supported for device type and device model. Filter strings
// are specified by the format "key=value,key=value". The filter
//     "type=temperature,model=ABC123"
// would only match devices whose type was temperature and model was ABC123.
func (p *Plugin) RegisterDeviceSetupActions(filter string, actions ...deviceAction) {
	if p.deviceSetupActions == nil {
		p.deviceSetupActions = make(map[string][]deviceAction)
	}
	if _, exists := p.deviceSetupActions[filter]; exists {
		p.deviceSetupActions[filter] = append(p.deviceSetupActions[filter], actions...)
	} else {
		p.deviceSetupActions[filter] = actions
	}
}

// Run starts the Plugin.
//
// Before the gRPC server is started, and before the read and write goroutines
// are started, Plugin setup and validation will happen. If successful, pre-run
// actions are executed, and device setup actions are executed, if defined.
func (p *Plugin) Run() error { // nolint: gocyclo
	logger.Info("Starting plugin run")
	err := p.setup()
	if err != nil {
		logger.Errorf("Failed plugin run setup: %v", err)
		return err
	}
	p.logInfo()

	// Before we start the dataManager goroutines or the gRPC server, we
	// will execute the preRunActions, if any exist.
	if len(p.preRunActions) > 0 {
		logger.Debug("Executing Pre Run Actions:")
		for _, action := range p.preRunActions {
			logger.Debugf(" * %v", action)
			err := action(p)
			if err != nil {
				logger.Errorf("Failed pre-run action %v: %v", action, err)
				return err
			}
		}
	}

	// At this point all state that the plugin will need should be available.
	// With a complete view of the plugin, devices, and configuration, we can
	// now process any device setup actions prior to reading to/writing from
	// the device(s).
	if len(p.deviceSetupActions) > 0 {
		logger.Debug("Executing Device Setup Actions:")
		for filter, actions := range p.deviceSetupActions {
			devices, err := filterDevices(filter)
			if err != nil {
				logger.Errorf("Failed to filter devices for setup actions: %v", err)
				return err
			}
			logger.Debugf("* %v (%v devices match filter %v)", actions, len(devices), filter)
			for _, d := range devices {
				for _, action := range actions {
					err := action(p, d)
					if err != nil {
						logger.Errorf("Failed device setup action %v: %v", action, err)
						return err
					}
				}
			}
		}
	}

	// Start the dataManager goroutines for reading and writing data.
	p.dataManager.init()

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
		logger.Errorf("Failed plugin handler validation: %v", err)
		return err
	}

	// validate that the plugin configuration is set
	if p.Config == nil {
		return fmt.Errorf("plugin must be configured before it is run")
	}

	// register configured devices with the plugin
	err = p.registerDevices()
	if err != nil {
		// we don't want to fail hard if registration fails, so just log out
		// the registration error instead.
		logger.Warnf("device registration failure: %v", err)
	}

	// Setup the transaction cache
	ttl, err := p.Config.Settings.Transaction.GetTTL()
	if err != nil {
		logger.Errorf("Bad transaction TTL config %v: %v", p.Config.Settings.Transaction.TTL, err)
		return err
	}
	err = setupTransactionCache(ttl)
	if err != nil {
		logger.Errorf("Failed to setup transaction cache: %v", err)
		return err
	}

	// Register a new server and dataManager for the Plugin. This should
	// be done prior to running the plugin, as opposed to on initialization
	// of the Plugin struct, because their configuration is configuration
	// dependent. The Plugin should be configured prior to running.
	p.server, err = newServer(p)
	if err != nil {
		logger.Errorf("Failed to create new gRPC server: %v", err)
		return err
	}

	// Create the dataManager
	p.dataManager, err = newDataManager(p)
	if err != nil {
		logger.Errorf("Failed to create plugin data manager: %v", err)
	}
	return err
}

// logInfo logs out the information about the plugin. This is called just before the
// plugin begins running all of its components.
func (p *Plugin) logInfo() {

	// Log out the plugin metainfo
	metainfo.log()

	// Log out the version info
	versionInfo := GetVersion()
	logger.Info(versionInfo.Format())

	// Log out configuration info
	logger.Infof("Plugin Config:")
	s, _ := json.MarshalIndent(p.Config, "", "  ")
	logger.InfoMultiline(string(s))
	logger.Info("Registered Devices:")
	for id, dev := range deviceMap {
		logger.Infof(" %v (%v)", id, dev.Model)
	}
	logger.Info("--------------------------------")
}
