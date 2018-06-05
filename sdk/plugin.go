package sdk

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
	"github.com/vapor-ware/synse-sdk/sdk/policies"
)

// outputTypeMap is a map where the the key is the name of the output type
// and the value is the corresponding OutputType.
var outputTypeMap = map[string]*config.OutputType{}

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

// A Plugin represents an instance of a Synse Plugin. Synse Plugins are used
// as data providers and device controllers for Synse Server.
type Plugin struct {
	policies []policies.ConfigPolicy

	server *Server
	quit   chan os.Signal
}

// NewPlugin creates a new instance of a Synse Plugin.
func NewPlugin(options ...PluginOption) *Plugin {
	plugin := Plugin{
		quit: make(chan os.Signal),
	}

	// Set custom options for the plugin.
	for _, option := range options {
		option(Context)
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

// RegisterOutputTypes registers OutputType instances with the Plugin. If a plugin
// is able to define its output types statically, they would be registered with the
// plugin via this method. Output types can also be registered via configuration
// file.
func (plugin *Plugin) RegisterOutputTypes(types ...*config.OutputType) error {
	multiErr := errors.NewMultiError("registering output types")
	logger.Debug("registering output types")
	for _, outputType := range types {
		_, hasType := outputTypeMap[outputType.Name]
		if hasType {
			multiErr.Add(fmt.Errorf("output type with name '%s' already exists", outputType.Name))
			continue
		}
		logger.Debugf("adding type: %s", outputType.Name)
		outputTypeMap[outputType.Name] = outputType
	}
	return multiErr.Err()
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
// filtering is only supported for device kind. Filter strings are specified in
// the format "key=value,key=value". The filter
//     "kind=temperature,kind=ABC123"
// would only match devices whose kind was temperature or ABC123.
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
func (plugin *Plugin) Run() (err error) {
	// Register system calls for graceful stopping.
	signal.Notify(plugin.quit, syscall.SIGTERM)
	signal.Notify(plugin.quit, syscall.SIGINT)
	go plugin.Stop()

	// The plugin name must be set as metainfo, since it is used in the Device
	// model. Check if it is set here. If not, return an error.
	if metainfo.Name == "" {
		return fmt.Errorf("plugin name not set, but required; see sdk.SetPluginMetainfo")
	}

	// ** "Config" steps **

	// Check for command line flags. If any flags are set that require an
	// action, that action will be resolved here.
	plugin.resolveFlags()

	// Check for configuration policies. If no policy was set by the plugin,
	// this will fall back on the default policies.
	err = plugin.checkPolicies()
	if err != nil {
		return
	}

	// Read in all configs and verify that they are correct.
	err = plugin.processConfig()
	if err != nil {
		return
	}

	// ** "Registration" steps **

	// Initialize Device instances for each of the devices configured with
	// the plugin.
	err = plugin.registerDevices()
	if err != nil {
		return
	}

	// ** "Making" steps **

	// Set up the transaction cache
	ttl, err := PluginConfig.Settings.Transaction.GetTTL()
	if err != nil {
		return
	}
	setupTransactionCache(ttl)

	// Initialize a gRPC server for the Plugin to use.
	plugin.server = NewServer(
		PluginConfig.Network.Type,
		PluginConfig.Network.Address,
	)

	// ** "Action" steps **

	// Before we start the dataManager goroutines or the gRPC server, we
	// will execute the preRunActions, if any exist.
	multiErr := execPreRun(plugin)
	if multiErr.HasErrors() {
		return multiErr
	}

	// At this point all state that the plugin will need should be available.
	// With a complete view of the plugin, devices, and configuration, we can
	// now process any device setup actions prior to reading to/writing from
	// the device(s).
	multiErr = execDeviceSetup(plugin)
	if multiErr.HasErrors() {
		return multiErr
	}

	// Log info at plugin startup
	plugin.logStartupInfo()

	// If the --dry-run flag is set, we will end here. The gRPC server and
	// data manager do not get started up in the dry run.
	if flagDryRun {
		logger.Info("dry-run successful")
		os.Exit(0)
	}

	logger.Debug("starting plugin server and manager")

	// "Starting" steps **

	// startDataManager
	DataManager.init()

	// startServer
	return plugin.server.Serve()
}

func (plugin *Plugin) Stop() {
	sig := <-plugin.quit
	logger.Infof("Stopping plugin (%s)...", sig.String())

	// TODO: any other stop/cleanup actions should go here (closing channels, etc)

	multiErr := execPostRun(plugin)
	if multiErr.HasErrors() {
		logger.Error(multiErr)
		os.Exit(1)
	}

	logger.Info("[done]")
	os.Exit(0)
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

	if flagDebug {
		logger.SetLogLevel(true)
	}

	// Print out the version info for the plugin.
	if flagVersion {
		fmt.Println(Version.Format())
		os.Exit(0)
	}
}

// checkPolicies checks for policies registered with the plugin. If no policies
// were set, the defaults will be used.
func (plugin *Plugin) checkPolicies() error {
	// Verify that the policies set for the plugin do not break any of the
	// constraints set on the policies.
	err := policies.CheckConstraints(plugin.policies)
	if err.Err() != nil {
		logger.Error("config policies set for the plugin are invalid")
		return err
	}

	// If we passed constraint checking, we can use the given policies
	policies.Set(plugin.policies)
	return nil
}

// processConfig handles plugin configuration in a number of steps. The behavior
// of config handling is dependent on the config policy that is set. If no config
// policies are set, the plugin will terminate in error.
//
// There are four major steps to processing plugin configuration: reading in the
// config, validating the config scheme, config unification, and verifying the
// config data is correct. These steps should happen for all config types.
func (plugin *Plugin) processConfig() error {

	// First, resolve the plugin config. We need to do this first, since subsequent
	// steps may require a plugin config to be specified.
	pluginPolicy := policies.PolicyManager.GetPluginConfigPolicy()
	logger.Debugf("plugin config policy: %s", pluginPolicy.String())

	pluginCtx, err := config.GetPluginConfigFromFile()
	if err != nil {
		// If we got an error when looking for the plugin config, we'll need to
		// check what the policy is for plugin config to determine how to proceed.
		switch pluginPolicy {
		case policies.PluginConfigRequired:
			// TODO: this should be a custom error
			return fmt.Errorf("policy violation: plugin config required but not found")
		case policies.PluginConfigOptional:
			// If the Plugin Config is optional, we will still need to create a new
			// plugin config that has all of the defaults filled out.
			cfg, err := config.NewDefaultPluginConfig()
			if err != nil {
				return err
			}
			PluginConfig = cfg
		default:
			return fmt.Errorf("unsupported plugin config policy: %s", pluginPolicy.String())
		}
	}

	multiErr := config.Validator.Validate(pluginCtx, pluginCtx.Source)
	if multiErr.HasErrors() {
		return multiErr
	}

	// If we get here, the plugin config was verified correctly. Assign it to the
	// global plugin config.
	PluginConfig = pluginCtx.Config.(*config.PluginConfig)

	// Now, resolve the other configs

	// Register output type configs from file
	outputTypeCtxs, err := config.GetOutputTypeConfigsFromFile()
	if err != nil {
		logger.Debug(err)
		if len(outputTypeMap) == 0 {
			// If we do not have any types already registered in the type map
			// (e.g. via plugin.RegisterOutputTypes), then we will have to fail
			// here.. if we don't know any output types, we won't be able to output
			// anything properly.
			return err
		}
	} else {
		for _, ctx := range outputTypeCtxs {
			cfg := ctx.Config.(*config.OutputType)
			err := plugin.RegisterOutputTypes(cfg)
			if err != nil {
				return err
			}
		}
	}
	// TODO: policies for type configs?

	// Resolve device configs both from file and from dynamic registration, if configured.
	deviceMultiErr := errors.NewMultiError("parsing device configs")
	deviceCtxs, err := config.GetDeviceConfigsFromFile()
	if err != nil {
		deviceMultiErr.Add(err)
	}

	// Get device config from dynamic registration, if anything is set there.
	// cfg schemes not validated yet, so not populated w/ default values...
	deviceConfigs, err := Context.dynamicDeviceConfigRegistrar(PluginConfig.DynamicRegistration.Config)
	if err != nil {
		deviceMultiErr.Add(err)
	}

	// If any device configs were found during dynamic registration, wrap them in
	// a context and add them to the known deviceCtxs.
	for _, cfg := range deviceConfigs {
		ctx := config.NewConfigContext("dynamic registration", cfg)
		deviceCtxs = append(deviceCtxs, ctx)
	}

	// FIXME: should there be different policies here.. e.g. config from file is optional
	// vs config entirely missing is bad?
	devicePolicy := policies.PolicyManager.GetDeviceConfigPolicy()
	logger.Debugf("device config policy: %s", devicePolicy.String())

	if deviceMultiErr.Err() != nil || len(deviceCtxs) == 0 {
		switch devicePolicy {
		case policies.DeviceConfigRequired:
			if deviceMultiErr.Err() != nil {
				logger.Error(deviceMultiErr)
			}
			// TODO this should be a custom error
			return fmt.Errorf("policy violation: device config(s) required, but none found")
		case policies.DeviceConfigOptional:
			// If the device config is optional, we should be fine without having found
			// anything at this point. We will add a default, empty device config to the
			// device config contexts so we can pass validation below.
			ctx := config.ConfigContext{
				Source: "default",
				Config: config.NewDeviceConfig(),
			}
			deviceCtxs = append(deviceCtxs, &ctx)

		default:
			return fmt.Errorf("unsupported device config policy: %s", devicePolicy.String())
		}
	}

	// 2. Validate Config Schemes
	for _, ctx := range deviceCtxs {
		multiErr = config.Validator.Validate(ctx, ctx.Source)
		if multiErr.HasErrors() {
			return multiErr
		}
	}

	// 3. Unify Configs
	unifiedCtx, err := config.UnifyDeviceConfigs(deviceCtxs)
	if err != nil {
		return err
	}

	// 4. Verify
	if !unifiedCtx.IsDeviceConfig() {
		return fmt.Errorf("unexpected config type for unified device configs: %v", unifiedCtx)
	}
	unifiedCfg := unifiedCtx.Config.(*config.DeviceConfig)
	multiErr = VerifyConfigs(unifiedCfg)
	if multiErr.HasErrors() {
		return multiErr
	}

	// If we are all set here, we the unified config should be our global device config
	DeviceConfig = unifiedCfg

	return nil
}

// registerDevices registers devices with the plugin. Devices are registered
// from multiple configuration sources: from file and from any dynamic registration
// functions supplied by the plugin.
//
// In both cases, the config sources are converted into the SDK's Device instances
// which represent the physical/virtual devices that the plugin will manage.
func (plugin *Plugin) registerDevices() error {

	// devices from dynamic registration
	devices, err := Context.dynamicDeviceRegistrar(PluginConfig.DynamicRegistration.Config)
	if err != nil {
		return err
	}
	updateDeviceMap(devices)

	// devices from config. the config here is the unified device config which
	// is joined from file and from dynamic registration, if set.
	devices, err = makeDevices(DeviceConfig)
	if err != nil {
		return err
	}
	updateDeviceMap(devices)

	return nil
}

// logStartupInfo is used to log plugin info at plugin startup. This will log
// the plugin metadata, version info, and registered devices.
func (plugin *Plugin) logStartupInfo() {
	// Log plugin metadata
	metainfo.log()

	// Log plugin version info
	Version.Log()

	// Log registered devices
	logger.Info("Registered Devices:")
	for id, dev := range deviceMap {
		logger.Infof("  %v (%v)", id, dev.Kind)
	}
	logger.Info("--------------------------------")
}
