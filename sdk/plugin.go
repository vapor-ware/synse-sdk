package sdk

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"

	// TODO: "config" is in the package namespace.. we'll need to clean
	//  that up so we don't need to alias the import
	cfg "github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/health"
	"github.com/vapor-ware/synse-sdk/sdk/policies"
)

const (
	// PluginEnvOverride defines the environment variable that can be used to
	// set an override config location for the Plugin configuration file.
	PluginEnvOverride = "PLUGIN_CONFIG"
)

var (
	flagDebug   bool
	flagVersion bool
	flagInfo    bool
	flagDryRun  bool
)

func init() {
	flag.BoolVar(&flagDebug, "debug", false, "enable debug logging")
	flag.BoolVar(&flagVersion, "version", false, "print the plugin version information")
	flag.BoolVar(&flagInfo, "info", false, "print the plugin metadata")
	flag.BoolVar(&flagDryRun, "dry-run", false, "run only the setup actions to verify functionality and configuration")

	// Logging defaults: set the level to info and use a formatter that gives
	// us millisecond resolution.
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.999Z07:00",
	})
}

// PluginAction defines an action that can be run before or after the main
// Plugin run logic. This is generally used for setup/teardown.
type PluginAction struct {
	Name string
	Action func(p *Plugin) error
}

// A Plugin represents an instance of a Synse Plugin. Synse Plugins are used
// as data providers and device controllers for Synse server.
type Plugin struct {
	quit   chan os.Signal
	config *cfg.Plugin

	server *server

	preRun  []*PluginAction
	postRun []*PluginAction
}

// NewPlugin creates a new instance of a Synse Plugin.
func NewPlugin(options ...PluginOption) (*Plugin, error) {
	// Normally, this would be a weird place to call Parse and do other config
	// setup, but since this constructor is effectively acting as the entry point
	// to the SDK, it works here.
	flag.Parse()

	// Prior to doing any other setup/loading, check if we are set to run
	// in debug mode.
	if flagDebug {
		log.SetLevel(log.DebugLevel)
	}

	// Load the plugin configuration.
	conf := new(cfg.Plugin)
	if err := loadPluginConfig(conf); err != nil {
		return nil, err
	}

	// If debug isn't set via command-line override and it is set in the
	// configuration file, set the level to debug, otherwise, keep it at info.
	if !flagDebug && conf.Debug {
		log.SetLevel(log.DebugLevel)
	}

	// Create the server used for gRPC communication. This will fail if the
	// server can not be set up (e.g. misconfiguration).
	server, err := newServer(conf.Network)
	if err != nil {
		return nil, err
	}

	// Create a new instance of the plugin.
	plugin := Plugin{
		quit:   make(chan os.Signal),
		config: conf,
		server: server,
	}

	// Register system calls for graceful stopping.
	signal.Notify(plugin.quit, syscall.SIGTERM)
	signal.Notify(plugin.quit, syscall.SIGINT)

	// Set custom options for the plugin.
	for _, option := range options {
		option(ctx)
	}
	return &plugin, nil
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

// RegisterPreRunActions registers actions with the Plugin which will be called prior
// to the business logic of the Plugin.
//
// Pre-run actions are considered setup actions / validator and as such, they are
// included in the Plugin dry-run. These run before the final pre-flight checks.
func (plugin *Plugin) RegisterPreRunActions(actions ...*PluginAction) {
	plugin.preRun = append(plugin.preRun, actions...)
}

// RegisterPostRunActions registers actions with the Plugin which will be called
// after it terminates.
//
// These actions are generally cleanup and teardown actions.
func (plugin *Plugin) RegisterPostRunActions(actions ...*PluginAction) {
	plugin.postRun = append(plugin.postRun, actions...)
}

//// RegisterDeviceSetupActions registers functions with the plugin that will be
//// called on device initialization before it is ever read from / written to. The
//// functions here can be used for device-specific setup actions.
////
//// The filter parameter should be the filter to apply to devices. Currently
//// filtering is supported for device kind and type. Filter strings are specified in
//// the format "key=value,key=value". The filter
////     "kind=temperature,kind=ABC123"
//// would only match devices whose kind was temperature or ABC123.
//func (plugin *Plugin) RegisterDeviceSetupActions(filter string, actions ...deviceAction) {
//	if _, exists := ctx.deviceSetupActions[filter]; exists {
//		ctx.deviceSetupActions[filter] = append(ctx.deviceSetupActions[filter], actions...)
//	} else {
//		ctx.deviceSetupActions[filter] = actions
//	}
//}

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

	// Before anything else is done, check to see if any command-line flags
	// were set which would terminate the run (e.g. printing version info
	// or plugin metadata).
	preRunPrint()

	// Run the pre-flight checks. If any check fails, we cannot run.
	if err := plugin.preFlightChecks(); err != nil {
		log.Error("[plugin] failed pre-flight check(s) - terminating")
		return err
	}

	// Perform pre-run setup
	if err := plugin.setup(); err != nil {
		return err
	}

	// Before we start the dataManager goroutines or the gRPC server, we
	// will execute the preRunActions, if any exist.
	if err := plugin.execPreRun(); err != nil {
		return err
	}

	//// At this point all state that the plugin will need should be available.
	//// With a complete view of the plugin, devices, and configuration, we can
	//// now process any device setup actions prior to reading to/writing from
	//// the device(s).
	//multiErr = execDeviceSetup(plugin)
	//if multiErr.HasErrors() {
	//	return multiErr
	//}

	// Log info at plugin startup
	logStartupInfo()

	// If the --dry-run flag is set, we will end here. The gRPC server and
	// data manager do not get started up in the dry run.
	if flagDryRun {
		log.Info("dry-run successful")
		os.Exit(0)
	}

	log.Debug("[sdk] starting plugin run")

	// todo: this should go somewhere here:
	// Listen for signals to terminate the Plugin.
	go plugin.onQuit()

	// If the default health checks are enabled, register them now
	if !plugin.config.Health.Checks.DisableDefaults {
		log.Debug("[sdk] registering default health checks")
		health.RegisterPeriodicCheck("read buffer health", 30*time.Second, readBufferHealthCheck)
		health.RegisterPeriodicCheck("write buffer health", 30*time.Second, writeBufferHealthCheck)
	}

	// Start the data manager
	if err := DataManager.run(); err != nil {
		return err
	}

	// Start the gRPC server
	return plugin.server.Serve()
}

// loadPluginConfig loads plugin configurations from file and environment
// and marshals that data into the provided Plugin config struct.
func loadPluginConfig(conf *cfg.Plugin) error {
	// Setup the config loader for the plugin.
	loader := cfg.NewYamlLoader("plugin")
	loader.EnvPrefix = "PLUGIN"
	loader.EnvOverride = PluginEnvOverride
	loader.FileName = "config"
	loader.AddSearchPaths(
		".",                        // Current working directory
		"./config",                 // Local config override directory
		"/etc/synse/plugin/config", // Default plugin config directory
	)

	// Load the plugin configuration.
	if err := loader.Load(); err != nil {
		return err
	}

	// Marshal the configuration into the plugin config struct.
	return loader.Scan(conf)
}

// preFlightChecks runs checks prior to starting the Plugin to ensure that it has
// all of the information it needs defined and available. Any failure here means
// that the Plugin will not be able to run.
func (plugin *Plugin) preFlightChecks() error {
	logOk := log.WithField("status", "ok")
	logErr := log.WithField("status", "failed")

	var failed bool

	// Check that the plugin metadata is set; only the name is required.
	// fixme (etd): there is probably a way to simplify this
	if metainfo.Name == "" {
		logErr.Error("[plugin] pre-flight: plugin name set")
		failed = true
	} else {
		logOk.Info("[plugin] pre-flight: plugin name set")
	}

	if failed {
		// fixme : custom pre-flight error?
		return fmt.Errorf("preflight checks failed")
	}
	return nil
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
	if err := plugin.execPostRun(); err != nil {
		log.Error(err)
		os.Exit(1)
	}

	log.Info("[done]")
	os.Exit(0)
}

// execPreRun executes the pre-run actions for the plugin.
func (plugin *Plugin) execPreRun() error {
	if len(plugin.preRun) == 0 {
		return nil
	}

	var multiErr = errors.NewMultiError("Pre-Run Actions")

	log.WithFields(log.Fields{
		"actions": len(plugin.preRun),
	}).Info("[plugin] executing pre-run actions")

	for _, action := range plugin.preRun {
		actionLog := log.WithField("action", action.Name)
		actionLog.Debug("[plugin] running pre-run action")
		if err := action.Action(plugin); err != nil {
			actionLog.Error("[plugin] pre-run action failed")
			multiErr.Add(err)
		}
	}
	return multiErr.Err()
}

// execPostRun executes the post-run actions for the plugin.
func (plugin *Plugin) execPostRun() error {
	if len(plugin.postRun) == 0 {
		return nil
	}

	var multiErr = errors.NewMultiError("Post-Run Actions")

	log.WithFields(log.Fields{
		"actions": len(plugin.postRun),
	}).Info("[plugin] executing post-run actions")

	for _, action := range plugin.postRun {
		actionLog := log.WithField("action", action.Name)
		actionLog.Debug("[plugin] running post-run action")
		if err := action.Action(plugin); err != nil {
			actionLog.Error("[plugin] post-run action failed")
			multiErr.Add(err)
		}
	}
	return multiErr.Err()
}

// setup performs the pre-run setup actions for a plugin.
func (plugin *Plugin) setup() error {
	// Register system calls for graceful stopping.
	//signal.Notify(plugin.quit, syscall.SIGTERM)
	//signal.Notify(plugin.quit, syscall.SIGINT)
	//go plugin.onQuit()

	//// The plugin name must be set as metainfo, since it is used in the Device
	//// model. Check if it is set here. If not, return an error.
	//if metainfo.Name == "" {
	//	return fmt.Errorf("plugin name not set, but required; see sdk.SetPluginMetainfo")
	//}

	// Check that the registered device handlers do not have any conflicting names.
	err := ctx.checkDeviceHandlers()
	if err != nil {
		return err
	}

	// Check for configuration policies. If no policy was set by the plugin,
	// this will fall back on the default policies.
	err = policies.Check()
	if err != nil {
		return err
	}

	//// Read in all configs and verify that they are correct.
	//err = plugin.processConfig()
	//if err != nil {
	//	return err
	//}

	//// If the plugin config specifies debug mode, enable debug mode
	//if plugin.config.Debug {
	//	log.SetLevel(log.DebugLevel)
	//}

	// Initialize Device instances for each of the devices configured with
	// the plugin.
	err = registerDevices()
	if err != nil {
		return err
	}

	// Set up the transaction cache
	setupTransactionCache(plugin.config.Settings.Transaction.TTL)

	// Set up the readings cache, if its configured
	setupReadingsCache()

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

	//// Resolve the plugin config.
	//log.Debug("[sdk] resolving plugin config")
	//err := processPluginConfig()
	//if err != nil {
	//	return err
	//}

	//// Resolve the output type config(s).
	//log.Debug("[sdk] resolving output type config(s)")
	//outputTypes, err := processOutputTypeConfig()
	//if err != nil {
	//	return err
	//}
	//
	//// Register the found output types, if any.
	//for _, output := range outputTypes {
	//	err = plugin.RegisterOutputTypes(output)
	//	if err != nil {
	//		return err
	//	}
	//}

	// Finally, make sure that we have output types. If we
	// don't, return an error, since we won't be able to properly
	// register devices.
	// todo: update output type registration
	if len(ctx.outputTypes) == 0 {
		return fmt.Errorf(
			"no output types found. you must either register output types " +
				"with the plugin, or configure them via file",
		)
	}

	// Resolve the device config(s).
	// todo: this should be done elsewhere (device manager?)
	log.Debug("[sdk] resolving device config(s)")
	err := processDeviceConfigs()
	if err != nil {
		return err
	}

	log.Debug("[sdk] finished processing configuration(s) for run")
	return nil
}

// preRunPrint prints out information about the plugin prior to doing any setup
// or run actions.
//
// If the item being printed is a command line option, it will terminate the
// plugin after printing.
func preRunPrint() {
	var terminate bool

	// --info was set; print the plugin metadata.
	if flagInfo {
		fmt.Println(metainfo.Format())
	}

	// --version was set; print the plugin version.
	if flagVersion {
		fmt.Println(version.Format())
	}

	if terminate {
		// fixme: for testing, should we use an Exiter interface?
		os.Exit(0)
	}
}
