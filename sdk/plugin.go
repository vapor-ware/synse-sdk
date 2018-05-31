package sdk

import (
	"flag"
	"fmt"
	"os"

	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
	"github.com/vapor-ware/synse-sdk/sdk/policies"
)

// DeviceIdentifier is a function that produces a string that can be used to
// identify a device deterministically. The returned string should be a composite
// from the Device's config data.
type DeviceIdentifier func(map[string]interface{}) string

// DynamicDeviceRegistrar is a function that takes a Plugin config's "dynamic
// registration" data and generates Device instances from it. How this is done
// is specific to the plugin/protocol.
type DynamicDeviceRegistrar func(map[string]interface{}) ([]*Device, error)

// DynamicDeviceConfigRegistrar is a function that takes a Plugin config's "dynamic
// registration" data and generates DeviceConfig instances from it. How this is done
// is specific to the plugin/protocol.
type DynamicDeviceConfigRegistrar func(map[string]interface{}) ([]*config.DeviceConfig, error)

/*

The tricky thing about plugins:

  We want to have a Plugin represent the thing doing the work and to have it
  hold the API for the user. At the same time, a lot of the data structures
  that we have should probably be global structures. This way, we do not have
  to constantly force things into the API/function sigs just so we can access
  a piece of data.

  With that in mind, all of this stuff really is tied to the plugin. There
  shouldn't be more than one plugin active at a time in a single project so
  using globals is okay, I think.

  What are some things that would be good to have as global?
    - the device map
	- the registered handlers list
	- the config
	- the transaction cache
	- the data manager
	- the plugin server?.. this may not need to be global and could be tied
	  directly to the plugin, I guess.

  So, the question becomes: How do we do all of this while maintaining a
  sane and useful API, and not making the underlying implementation too
  convoluted?

  We could just have the plugin be global................................
  If the plugin is global, all of the data that is associated with it should
  effectively be global as well. That seems kinda gross though.

*/

// A Plugin represents an instance of a Synse Plugin. Synse Plugins are used
// as data providers and device controllers for Synse Server.
type Plugin struct {
	policies []policies.ConfigPolicy

	deviceIdentifier             DeviceIdentifier
	dynamicDeviceRegistrar       DynamicDeviceRegistrar
	dynamicDeviceConfigRegistrar DynamicDeviceConfigRegistrar
}

// NewPlugin creates a new instance of a Synse Plugin.
func NewPlugin(options ...PluginOption) *Plugin {
	plugin := Plugin{}

	// Set custom options for the plugin.
	for _, option := range options {
		option(&plugin)
	}

	// Apply defaults to any required field that was not set from an option.
	for _, option := range defaultOptions {
		option(&plugin)
	}

	return &plugin
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
	preRunActions = append(preRunActions, actions...)
}

// RegisterPostRunActions registers functions with the plugin that will be called
// after the gRPC server and dataManager terminate running. The functions here can
// be used for plugin-wide teardown actions.
//
// NOTE: While post run actions can be defined for a Plugin, they are currently
// not executed. See: https://github.com/vapor-ware/synse-sdk/issues/85
func (plugin *Plugin) RegisterPostRunActions(actions ...pluginAction) {
	postRunActions = append(postRunActions, actions...)
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
	if _, exists := deviceSetupActions[filter]; exists {
		deviceSetupActions[filter] = append(deviceSetupActions[filter], actions...)
	} else {
		deviceSetupActions[filter] = actions
	}
}

// RegisterDeviceHandlers adds DeviceHandlers to the Plugin.
//
// These DeviceHandlers are then matched with the Device instances
// by their type/model and provide the read/write functionality for the
// Devices. If a DeviceHandler for a Device is not registered here, the
// Device will not be usable by the plugin.
func (plugin *Plugin) RegisterDeviceHandlers(handlers ...*DeviceHandler) {
	deviceHandlers = append(deviceHandlers, handlers...)
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
	execPreRun(plugin)

	// At this point all state that the plugin will need should be available.
	// With a complete view of the plugin, devices, and configuration, we can
	// now process any device setup actions prior to reading to/writing from
	// the device(s).
	execDeviceSetup(plugin)

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

	// Get device config from dynamic registration, if anything is set there.
	deviceConfigs, err := plugin.dynamicDeviceConfigRegistrar() // TODO: pass in correct param
	if err != nil {
		// todo: error
	}

	// If any device configs were found during dynamic registration, wrap them in
	// a context and add them to the known deviceCtxs.
	for _, cfg := range deviceConfigs {
		ctx := config.NewConfigContext("dynamic registration", cfg)
		deviceCtxs = append(deviceCtxs, ctx)
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

	// devices from dynamic registration
	devices, err := plugin.dynamicDeviceRegistrar() // TODO: pass in correct param
	if err != nil {
		// todo: error
	}

	// Update global devices map
	for _, d := range devices {
		if _, hasDevice := deviceMap[d.GUID()]; hasDevice {
			// todo: error -- we already have this device in the map...
		}
		deviceMap[d.GUID()] = d
	}

	// devices from config. the config here is the unified device config which
	// is joined from file and from dynamic registration, if set.
	devices, err = makeDevices(&DeviceConfig)
	if err != nil {
		// todo: error
	}

	// Update global devices map
	for _, d := range devices {
		if _, hasDevice := deviceMap[d.GUID()]; hasDevice {
			// todo: error -- we already have this device in the map...
		}
		deviceMap[d.GUID()] = d
	}
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
		logger.Infof("  %v (%v)", id, dev.Kind)
	}

	logger.Info("--------------------------------")
}
