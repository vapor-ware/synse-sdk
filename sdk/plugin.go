package sdk

import (
	"flag"
	"fmt"
	"os"

	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
	"github.com/vapor-ware/synse-sdk/sdk/policies"
)

/*
FIXME: this can be removed later, just adding as a note for now
According to the 1.0 Design Doc, we want there to be two ways of configuring a plugin.
 1. Simple Plugin
 2. Custom Plugin

A Simple Plugin would just use default handlers, and should be pretty easy to init, e.g.

  sdk.New()
or
  sdk.NewPlugin()

A Custom Plugin would take a little more work in that it would need to define some interfaces/
something in order to provide custom functionality. This could be done a few different ways.

1.) A set of interfaces.
  The plugin developer could define their own struct that fulfils some number of interfaces that
  provide the functionality for whatever they want. This isn't too bad and seems kinda nice, but
  internally, it gets kinda weird to parse all of this.

2.) A sdk.CustomPlugin function that takes functions as arguments.
  Those functions can be applied for the functionality needed, if specified. With this, the author
  would only have to define the functions. The downside being that this initializer could become
  really big and everything the author doesn't want a custom override for would have to be specified
  as nil... not great.

3.) A PluginOptions struct to define custom functionality
  I think this may be the best way of doing it.. sorta a middle ground between the other two?
  Basically, we can have something like

  sdk.NewPlugin(... options)

  Where an option would be the bits of configurable functionality? TBD how we define an option though,
  especially in this context. Perhaps it would make more sense to have a struct,

  options {
    DeviceIdentifier func()...
  }

  and then just have them pass in that struct? or something like that?


*/

// Plugin is the New Plugin. This will replace Plugin once v1.0 work is completed.
type Plugin struct {

	name string
	maintainer string
	description string
	vcs string



	policies []policies.ConfigPolicy

	preRunActions      []pluginAction
	postRunActions     []pluginAction
	deviceSetupActions map[string][]deviceAction
}

// NewPlugin creates a new instance of a Synse Plugin.
func NewPlugin(name, maintainer, description, vcs string) *Plugin {
	return &Plugin{
		name: name,
		maintainer: maintainer,
		description: description,
		vcs: vcs,
	}
}

// SetConfigPolicies sets the config policies for the plugin. Config policies will
// determine how the plugin behaves when reading in configurations.
//
// If no config policies are set, default policies will be used. Policy validation
// does not happen in this function. Config policies are validated on plugin Run.
func (plugin *Plugin) SetConfigPolicies(policies ...policies.ConfigPolicy) {
	plugin.policies = policies
}

// RegisterPreRunActions registers functions with the plugin that will be called
// before the gRPC server and dataManager are started. The functions here can be
// used for plugin-wide setup actions.
func (plugin *Plugin) RegisterPreRunActions(actions ...pluginAction) {
	if plugin.preRunActions == nil {
		plugin.preRunActions = actions
	} else {
		plugin.preRunActions = append(plugin.preRunActions, actions...)
	}
}

// RegisterPostRunActions registers functions with the plugin that will be called
// after the gRPC server and dataManager terminate running. The functions here can
// be used for plugin-wide teardown actions.
//
// NOTE: While post run actions can be defined for a Plugin, they are currently
// not executed. See: https://github.com/vapor-ware/synse-sdk/issues/85
func (plugin *Plugin) RegisterPostRunActions(actions ...pluginAction) {
	if plugin.postRunActions == nil {
		plugin.postRunActions = actions
	} else {
		plugin.postRunActions = append(plugin.postRunActions, actions...)
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
func (plugin *Plugin) RegisterDeviceSetupActions(filter string, actions ...deviceAction) {
	if plugin.deviceSetupActions == nil {
		plugin.deviceSetupActions = make(map[string][]deviceAction)
	}
	if _, exists := plugin.deviceSetupActions[filter]; exists {
		plugin.deviceSetupActions[filter] = append(plugin.deviceSetupActions[filter], actions...)
	} else {
		plugin.deviceSetupActions[filter] = actions
	}
}

// Run starts the Plugin.
//
// Before the gRPC server is started, and before the read and write goroutines
// are started, Plugin setup and validation will happen. If successful, pre-run
// actions are executed, and device setup actions are executed, if defined.
func (plugin *Plugin) Run() error {

	// ** "Config" steps **

	// Check for command line flags. If any flags are set that require an
	// action, that action will be resolved here.
	plugin.resolveFlags()

	// Check for configuration policies. If no policy was set by the plugin,
	// this will fall back on the default policies.
	plugin.checkPolicies()

	// Read in all configs and verify that they are correct.
	plugin.processConfig()

	// ** "Registration" steps **

	// FIXME: at this point, we will need to resolve dynamic registration stuff.
	//   - if dynamic registration makes Devices, just add to the global map.
	//   - if dynamic registration makes Config artifacts, add to the unified config.
	//     (or maybe this happens in th processConfig step, before unification?)
	// Initialize Device instances for each of the devices configured with
	// the plugin.
	plugin.registerDevices()

	// ** "Making" steps **

	// makeTransactionCache
	// ttl, err := p.Config.Settings.Transaction.GetTTL()
	// err = setupTransactionCache(ttl)

	// makeDataManager
	// makeServer

	// ** "Action" steps **

	// Before we start the dataManager goroutines or the gRPC server, we
	// will execute the preRunActions, if any exist.
	if len(plugin.preRunActions) > 0 {
		logger.Debug("Executing Pre Run Actions:")
		for _, action := range plugin.preRunActions {
			logger.Debugf(" * %v", action)
			err := action(plugin)
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
	if len(plugin.deviceSetupActions) > 0 {
		logger.Debug("Executing Device Setup Actions:")
		for filter, actions := range plugin.deviceSetupActions {
			devices, err := filterDevices(filter)
			if err != nil {
				logger.Errorf("Failed to filter devices for setup actions: %v", err)
				return err
			}
			logger.Debugf("* %v (%v devices match filter %v)", actions, len(devices), filter)
			for _, d := range devices {
				for _, action := range actions {
					err := action(plugin, d)
					if err != nil {
						logger.Errorf("Failed device setup action %v: %v", action, err)
						return err
					}
				}
			}
		}
	}

	// Log info at plugin startup
	plugin.logStartupInfo()

	// If the --dry-run flag is set, we will end here. The gRPC server and
	// data manager do not get started up in the dry run.
	if !flagDryRun {

		// "Starting" steps **

		// startDataManager
		// startServer

	}

	// profit

	return nil
}

// FIXME: its not clear that these private functions need to hang off the Plugin struct..

// resolveFlags parses flags passed to the plugin.
//
// Not all flags will result in an action, so not all flags are checked
// here. Only the flags that cause immediate actions will be processed here.
// Only SDK supported flags are parsed here. If a plugin specifies additional
// flags, they should be resolved in their own pre-run action.
func (plugin *Plugin) resolveFlags() {
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
func (plugin *Plugin) checkPolicies() {
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
func (plugin *Plugin) processConfig() {

	// FIXME: here, there are similar patterns for checking the policies..
	// perhaps this could live somewhere else?

	// 1. Read in the configs. We have a few types of configs to read in.
	//  a. Plugin Config
	pluginCtx, err := config.GetPluginConfigFromFile()
	if err != nil {
		switch policies.PolicyManager.GetPluginConfigPolicy() {
		case policies.PluginConfigOptional:
			// The plugin config is optional: do not fail if the config is not found.
		case policies.PluginConfigRequired:
			// The plugin config is required: fail if the config is not found.
		default:
			// log.Fatal? -- unsupported plugin config policiy
		}
		// what do we do when we error out here?
	}

	//  b. Device Config
	deviceCtxs, err := config.GetDeviceConfigsFromFile()
	if err != nil {
		switch policies.PolicyManager.GetDeviceConfigPolicy() {
		case policies.DeviceConfigOptional:
			// The device config is optional: do not fail if no configs are found.
		case policies.DeviceConfigRequired:
			// The device config is required: fail if no configs are found.
		default:
			// log.Fatal? -- unsupported device config policy
		}

		// what do we do when we error out here?
	}

	//  c. Type Config
	// TODO

	//  d. ... Other?
	// TODO?

	// FIXME: if no configs were found and we haven't failed yet (e.g. config policy
	// is optional), then we should not fail validation. i.e. the pluginCtx/deviceCtxs
	// should still be valid for all the steps below.

	// 2. Validate Config Schemes
	multiErr := config.Validator.Validate(pluginCtx, pluginCtx.Source)
	if multiErr.HasErrors() {
		// what to do here?
	}

	for _, ctx := range deviceCtxs {
		multiErr = config.Validator.Validate(ctx, ctx.Source)
		if multiErr.HasErrors() {
			// what to do here?
		}
	}

	// 3. Unify Configs
	unifiedCtx, err := config.UnifyDeviceConfigs(deviceCtxs)
	if err != nil {
		// what to do here?
	}

	// 4. Verify
	if !unifiedCtx.IsDeviceConfig() {
		// ERROR
	}
	unifiedCfg := unifiedCtx.Config.(*config.DeviceConfig)
	multiErr = config.VerifyConfigs(unifiedCfg)
	if multiErr.HasErrors() {
		// what to do here?
	}

	// TODO: once everything is done, what do we do with all the data?
}

// registerDevices registers devices with the plugin. Devices are registered
// from multiple configuration sources: from file and from any dynamic registration
// functions supplied by the plugin.
//
// In both cases, the config sources are converted into the SDK's Device instances
// which represent the physical/virtual devices that the plugin will manage.
func (plugin *Plugin) registerDevices() {

	// devices from config

	// devices from dynamic registration

}

// logStartupInfo is used to log plugin info at plugin startup. This will log
// the plugin metadata, version info, and registered devices.
func (plugin *Plugin) logStartupInfo() {
	// Log plugin metadata
	metainfo.log()

	// Log plugin version info
	version := GetVersion()
	logger.Info(version.Format())

	// Log registered devices
	logger.Info("Registered Devices:")
	for id, dev := range deviceMap {
		logger.Infof("  %v (%v)", id, dev.Model)
	}

	logger.Info("--------------------------------")
}

//
//// OPlugin represents an instance of a Synse plugin. Along with metadata
//// and definable handlers, it contains a gRPC server to handle the plugin
//// requests.
//type OPlugin struct {
//
//	deviceHandlers     []*DeviceHandler          // Plugin-specific read and write functions for the devices supported by the plugin.
//}
//
//
//// RegisterDeviceHandlers adds DeviceHandlers to the Plugin.
////
//// These DeviceHandlers are then matched with the Device instances
//// by their type/model and provide the read/write functionality for the
//// Devices. If a DeviceHandler for a Device is not registered here, the
//// Device will not be usable by the plugin.
//func (p *OPlugin) RegisterDeviceHandlers(handlers ...*DeviceHandler) {
//	if p.deviceHandlers == nil {
//		p.deviceHandlers = handlers
//	} else {
//		p.deviceHandlers = append(p.deviceHandlers, handlers...)
//	}
//}
//
//// registerDevices registers all of the configured devices (via their proto and
//// instance config) with the plugin.
//func (p *OPlugin) registerDevices() error {
//	var devices []*config.DeviceConfig
//
//	cfgDevices, err := devicesFromConfig()
//	if err != nil {
//		logger.Errorf("Failed to register devices from files: %v", err)
//		return err
//	}
//	devices = append(devices, cfgDevices...)
//
//	enumDevices, err := devicesFromAutoEnum(p)
//	if err != nil {
//		logger.Errorf("Failed to register devices from auto-enum: %v", err)
//		return err
//	}
//	devices = append(devices, enumDevices...)
//
//	return registerDevices(p, devices)
//}
