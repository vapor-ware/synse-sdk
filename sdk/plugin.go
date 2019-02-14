package sdk

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/health"
	"github.com/vapor-ware/synse-sdk/sdk/policies"
)

// A Plugin represents an instance of a Synse Plugin. Synse Plugins are used
// as data providers and device controllers for Synse server.
type Plugin struct {
	server *server
	quit   chan os.Signal
}

// NewPlugin creates a new instance of a Synse Plugin.
func NewPlugin(options ...PluginOption) *Plugin {
	plugin := Plugin{
		quit: make(chan os.Signal),
	}

	// Set custom options for the plugin.
	for _, option := range options {
		option(ctx)
	}
	return &plugin
}

// RegisterOutputTypes registers OutputType instances with the Plugin. If a plugin
// is able to define its output types statically, they would be registered with the
// plugin via this method. Output types can also be registered via configuration
// file.
func (plugin *Plugin) RegisterOutputTypes(types ...*OutputType) error {
	multiErr := errors.NewMultiError("registering output types")
	log.Debug("[sdk] registering output types")
	for _, outputType := range types {
		_, hasType := ctx.outputTypes[outputType.Name]
		if hasType {
			log.WithField("type", outputType.Name).Error("[sdk] output type already exists")
			multiErr.Add(fmt.Errorf("output type with name '%s' already exists", outputType.Name))
			continue
		}
		log.WithField("type", outputType.Name).Debug("[sdk] adding new output type")
		ctx.outputTypes[outputType.Name] = outputType
	}
	return multiErr.Err()
}

// RegisterPreRunActions registers functions with the plugin that will be called
// before the gRPC server and dataManager are started. The functions here can be
// used for plugin-wide setup actions.
func (plugin *Plugin) RegisterPreRunActions(actions ...pluginAction) {
	ctx.preRunActions = append(ctx.preRunActions, actions...)
}

// RegisterPostRunActions registers functions with the plugin that will be called
// after the gRPC server and dataManager terminate running. The functions here can
// be used for plugin-wide teardown actions.
func (plugin *Plugin) RegisterPostRunActions(actions ...pluginAction) {
	ctx.postRunActions = append(ctx.postRunActions, actions...)
}

// RegisterDeviceSetupActions registers functions with the plugin that will be
// called on device initialization before it is ever read from / written to. The
// functions here can be used for device-specific setup actions.
//
// The filter parameter should be the filter to apply to devices. Currently
// filtering is supported for device kind and type. Filter strings are specified in
// the format "key=value,key=value". The filter
//     "kind=temperature,kind=ABC123"
// would only match devices whose kind was temperature or ABC123.
func (plugin *Plugin) RegisterDeviceSetupActions(filter string, actions ...deviceAction) {
	if _, exists := ctx.deviceSetupActions[filter]; exists {
		ctx.deviceSetupActions[filter] = append(ctx.deviceSetupActions[filter], actions...)
	} else {
		ctx.deviceSetupActions[filter] = actions
	}
}

// RegisterDeviceHandlers adds DeviceHandlers to the Plugin.
//
// These DeviceHandlers are then matched with the Device instances
// by their name and provide the read/write functionality for the
// Devices. If a DeviceHandler is not registered for a Device, the
// Device will not be usable by the plugin.
func (plugin *Plugin) RegisterDeviceHandlers(handlers ...*DeviceHandler) {
	ctx.deviceHandlers = append(ctx.deviceHandlers, handlers...)
}

// Run starts the Plugin.
//
// Before the gRPC server is started, and before the read and write goroutines
// are started, Plugin setup and validation will happen. If successful, pre-run
// actions are executed, and device setup actions are executed, if defined.
func (plugin *Plugin) Run() error {
	// Perform pre-run setup
	err := plugin.setup()
	if err != nil {
		return err
	}

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
	logStartupInfo()

	// If the --dry-run flag is set, we will end here. The gRPC server and
	// data manager do not get started up in the dry run.
	if flagDryRun {
		log.Info("dry-run successful")
		os.Exit(0)
	}

	log.Debug("[sdk] starting plugin run")

	// If the default health checks are enabled, register them now
	if Config.Plugin.Health.UseDefaults {
		log.Debug("[sdk] registering default health checks")
		health.RegisterPeriodicCheck("read buffer health", 30*time.Second, readBufferHealthCheck)
		health.RegisterPeriodicCheck("write buffer health", 30*time.Second, writeBufferHealthCheck)
	}

	// Start the data manager
	err = DataManager.run()
	if err != nil {
		return err
	}

	// Start the gRPC server
	return plugin.server.Serve()
}

// onQuit is a function that waits for a signal to terminate the plugin's run
// and run cleanup/post-run actions prior to terminating.
func (plugin *Plugin) onQuit() {
	sig := <-plugin.quit
	log.Infof("[sdk] stopping plugin (%s)...", sig.String())

	// TODO: any other stop/cleanup actions should go here (closing channels, etc)

	// Immediately stop the gRPC server.
	plugin.server.Stop()

	// Execute post-run actions.
	multiErr := execPostRun(plugin)
	if multiErr.HasErrors() {
		log.Error(multiErr)
		os.Exit(1)
	}

	log.Info("[done]")
	os.Exit(0)
}

// setupLogger sets up the logger. Currently this just gives us sub second time
// resolution, but can be expanded on later.
func setupLogger() error {
	// Set formatter that gives at least milliseconds.
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.999Z07:00",
	})

	return nil // There may be scenarios where we need to fail later (unclear).
}

// setup performs the pre-run setup actions for a plugin.
func (plugin *Plugin) setup() error {
	// Register system calls for graceful stopping.
	signal.Notify(plugin.quit, syscall.SIGTERM)
	signal.Notify(plugin.quit, syscall.SIGINT)
	go plugin.onQuit()

	err := setupLogger()
	if err != nil {
		return err
	}

	// The plugin name must be set as metainfo, since it is used in the Device
	// model. Check if it is set here. If not, return an error.
	if metainfo.Name == "" {
		return fmt.Errorf("plugin name not set, but required; see sdk.SetPluginMetainfo")
	}

	// Check for command line flags. If any flags are set that require an
	// action, that action will be resolved here.
	parseFlags()

	// Check that the registered device handlers do not have any conflicting names.
	err = ctx.checkDeviceHandlers()
	if err != nil {
		return err
	}

	// Check for configuration policies. If no policy was set by the plugin,
	// this will fall back on the default policies.
	err = policies.Check()
	if err != nil {
		return err
	}

	// Read in all configs and verify that they are correct.
	err = plugin.processConfig()
	if err != nil {
		return err
	}

	// If the plugin config specifies debug mode, enable debug mode
	if Config.Plugin.Debug {
		log.SetLevel(log.DebugLevel)
	}

	// Initialize Device instances for each of the devices configured with
	// the plugin.
	err = registerDevices()
	if err != nil {
		return err
	}

	// Set up the transaction cache
	ttl, err := Config.Plugin.Settings.Transaction.GetTTL()
	if err != nil {
		return err
	}
	setupTransactionCache(ttl)

	// Set up the readings cache, if its configured
	setupReadingsCache()

	// Initialize a gRPC server for the Plugin to use.
	plugin.server = newServer(
		Config.Plugin.Network.Type,
		Config.Plugin.Network.Address,
	)
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

	// Resolve the plugin config.
	log.Debug("[sdk] resolving plugin config")
	err := processPluginConfig()
	if err != nil {
		return err
	}

	// Resolve the output type config(s).
	log.Debug("[sdk] resolving output type config(s)")
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
	if len(ctx.outputTypes) == 0 {
		return fmt.Errorf(
			"no output types found. you must either register output types " +
				"with the plugin, or configure them via file",
		)
	}

	// Resolve the device config(s).
	log.Debug("[sdk] resolving device config(s)")
	err = processDeviceConfigs()
	if err != nil {
		return err
	}

	log.Debug("[sdk] finished processing configuration(s) for run")
	return nil
}

//// The current (latest) version of the plugin config scheme.
//var currentPluginSchemeVersion = 3
//
//// NewDefaultPluginConfig creates a new instance of a PluginConfig with its
//// default values resolved.
//func NewDefaultPluginConfig() (*PluginConfig, error) {
//	config := &PluginConfig{
//		Version: currentPluginSchemeVersion,
//	}
//	err := defaults.Set(config)
//	if err != nil {
//		return nil, err
//	}
//	return config, nil
//}
