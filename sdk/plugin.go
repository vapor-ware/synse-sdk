package sdk

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/health"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
	"github.com/vapor-ware/synse-sdk/sdk/policies"
)

// outputTypeMap is a map where the the key is the name of the output type
// and the value is the corresponding OutputType.
var outputTypeMap = map[string]*config.OutputType{}

// A Plugin represents an instance of a Synse Plugin. Synse Plugins are used
// as data providers and device controllers for Synse Server.
type Plugin struct {
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
func (plugin *Plugin) Run() error { // nolint: gocyclo
	// Register system calls for graceful stopping.
	signal.Notify(plugin.quit, syscall.SIGTERM)
	signal.Notify(plugin.quit, syscall.SIGINT)
	go plugin.OnQuit()

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
	err := policies.Check()
	if err != nil {
		return err
	}

	// Read in all configs and verify that they are correct.
	err = plugin.processConfig()
	if err != nil {
		return err
	}

	// ** "Registration" steps **

	// Initialize Device instances for each of the devices configured with
	// the plugin.
	err = plugin.registerDevices()
	if err != nil {
		return err
	}

	// ** "Making" steps **

	// Set up the transaction cache
	ttl, err := PluginConfig.Settings.Transaction.GetTTL()
	if err != nil {
		return err
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

	// If the default health checks are enabled, register them now
	if PluginConfig.Health.UseDefaults {
		health.RegisterPeriodicCheck("read buffer health", 30*time.Second, readBufferHealthCheck)
		health.RegisterPeriodicCheck("write buffer health", 30*time.Second, writeBufferHealthCheck)
	}

	// startDataManager
	err = DataManager.run()
	if err != nil {
		return err
	}

	// startServer
	return plugin.server.Serve()
}

// OnQuit is a function that waits for a signal to terminate the plugin's run
// and run cleanup/post-run actions prior to terminating.
func (plugin *Plugin) OnQuit() {
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

// processConfig handles plugin configuration in a number of steps. The behavior
// of config handling is dependent on the config policy that is set. If no config
// policies are set, the plugin will terminate in error.
//
// There are four major steps to processing plugin configuration: reading in the
// config, validating the config scheme, config unification, and verifying the
// config data is correct. These steps should happen for all config types.
func (plugin *Plugin) processConfig() error {

	// Resolve the plugin config.
	err := processPluginConfig()
	if err != nil {
		return err
	}

	// Resolve the output type config(s).
	outputTypes, err := processOutputTypeConfig()
	if err != nil {
		return err
	}

	// Register the found output types, if any.
	for _, output := range outputTypes {
		err = plugin.RegisterOutputTypes(output)
		if err != nil {
			return err
		}
	}

	// Finally, make sure that we have output types. If we
	// don't, return an error, since we won't be able to properly
	// register devices.
	if len(outputTypeMap) == 0 {
		return fmt.Errorf(
			"no output types found. you must either register output types " +
				"with the plugin, or configure them via file",
		)
	}

	// Resolve the device config(s).
	err = processDeviceConfigs()
	if err != nil {
		return err
	}

	// Verify the unified config, and validate plugin-specific data.
	// FIXME: if we reorganize the SDK a bit, we can move the below to the above
	// function, but for now it needs to live here.
	multiErr := VerifyConfigs(config.Device)
	if multiErr.HasErrors() {
		return multiErr
	}
	multiErr = config.Device.ValidateDeviceConfigData(Context.deviceDataValidator)
	if multiErr.HasErrors() {
		return multiErr
	}

	logger.Debug("finished processing configuration(s) for run")
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
