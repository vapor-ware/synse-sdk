package sdk

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/vapor-ware/synse-sdk/sdk/output"

	log "github.com/Sirupsen/logrus"

	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
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
}

// PluginAction defines an action that can be run before or after the main
// Plugin run logic. This is generally used for setup/teardown.
type PluginAction struct {
	Name   string
	Action func(p *Plugin) error
}

// Plugin is a Synse Plugin.
type Plugin struct {
	info    *PluginMetadata
	version *pluginVersion
	config  *config.Plugin
	quit    chan os.Signal

	// Actions
	preRun  []*PluginAction
	postRun []*PluginAction

	// Plugin outputs
	outputs map[string]*output.Output

	// Options and handlers
	deviceIdentifier       DeviceIdentifier
	dynamicRegistrar       DynamicDeviceRegistrar
	dynamicConfigRegistrar DynamicDeviceConfigRegistrar
	deviceDataValidator    DeviceDataValidator

	pluginCfgRequired  bool
	deviceCfgRequired  bool
	dynamicCfgRequired bool

	// Plugin components
	scheduler     *Scheduler
	stateManager  *StateManager
	deviceManager *deviceManager
	server        *server
}

// NewPlugin creates a new instance of a Plugin. This should be the only
// way that a Plugin is initialized.
//
// This constructor will load the plugin configuration; if it is not present
// or invalid, this will fail. All other Plugin component initialization
// is deferred until Run is called.
func NewPlugin(options ...PluginOption) (*Plugin, error) {

	// Load the plugin configuration.
	conf := new(config.Plugin)
	if err := loadPluginConfig(conf); err != nil {
		return nil, err
	}

	meta := new(PluginMetadata)

	// Initialize plugin components.
	dm := newDeviceManager()
	sm := NewStateManager(conf.Settings)
	sched := NewScheduler(conf.Settings, dm, sm)
	server := newServer(conf.Network, dm, sm, sched, meta)

	p := Plugin{
		outputs: make(map[string]*output.Output),
		quit:    make(chan os.Signal),
		info:    meta,
		version: version,
		config:  conf,

		pluginCfgRequired:  false,
		deviceCfgRequired:  true,
		dynamicCfgRequired: false,

		deviceIdentifier:       defaultDeviceIdentifier,
		dynamicRegistrar:       defaultDynamicDeviceRegistration,
		dynamicConfigRegistrar: defaultDynamicDeviceConfigRegistration,
		deviceDataValidator:    defaultDeviceDataValidator,

		deviceManager: dm,
		stateManager:  sm,
		scheduler:     sched,
		server:        server,
	}

	// Set custom options for the plugin.
	for _, option := range options {
		option(&p)
	}

	// Register the built-in outputs with the plugin.
	if err := p.RegisterOutputs(output.GetBuiltins()...); err != nil {
		return nil, err
	}

	return &p, nil
}

// SetInfo sets the metadata for the Plugin. At a minimum, the Plugin name
// needs to be set.
func (plugin *Plugin) SetInfo(info *PluginMetadata) {
	plugin.info = info
}

// Run starts the plugin.
//
// This is the functional starting point for all plugins. Once this is called,
// the plugin will initialize all of its components and validate its state. Once
// everything is ready, it will run each of its components. The gRPC server is
// run in the foreground; all other components are run as goroutines.
func (plugin *Plugin) Run() error {
	// Check if anything needs to happen based on command line arguments.
	plugin.handleRunOptions()

	// Initialize the plugin and its components.
	if err := plugin.initialize(); err != nil {
		return err
	}

	// If all components initialized without error, we can register
	// any pre/post run actions which they may have.
	plugin.server.registerActions(plugin)
	plugin.scheduler.registerActions(plugin)

	// Run pre-run actions, if any exist.
	if err := plugin.execPreRun(); err != nil {
		return err
	}

	// Listen for plugin quit.
	go plugin.onQuitSignal()

	// If the plugin was run with the '--dry-run' flag, end the run here
	// before we actually start any of the plugin components.
	if flagDryRun {
		log.Info("[plugin] dry-run successful")
		os.Exit(0)
	}

	//// If the default health checks are enabled, register them now
	//// fixme (etd) - this will move to the health manager
	//if !plugin.config.Health.Checks.DisableDefaults {
	//	log.Debug("[sdk] registering default health checks")
	//	health.RegisterPeriodicCheck("read buffer health", 30*time.Second, readBufferHealthCheck)
	//	health.RegisterPeriodicCheck("write buffer health", 30*time.Second, writeBufferHealthCheck)
	//}

	// Run the plugin.
	return plugin.run()
}

// RegisterOutputs registers new Outputs with the Plugin. A plugin will automatically
// register the built-in SDK outputs. This function allows a plugin do augment that
// set of outputs with its own custom outputs.
//
// If any registered output names conflict with those of built-in or other custom
// outputs, an error is returned.
func (plugin *Plugin) RegisterOutputs(outputs ...*output.Output) error {
	multiErr := errors.NewMultiError("output registration")

	for _, o := range outputs {
		if _, exists := plugin.outputs[o.Name]; exists {
			multiErr.Add(fmt.Errorf("conflict: output with name '%s' already exists", o.Name))
			continue
		}
		plugin.outputs[o.Name] = o
	}
	return multiErr.Err()
}

// RegisterPreRunActions registers actions with the Plugin which will be called prior
// to the business logic of the Plugin.
//
// Pre-run actions are considered setup/validator actions and as such, they are
// included in the Plugin dry-run.
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

// RegisterDeviceHandlers adds DeviceHandlers to the Plugin.
//
// These DeviceHandlers are matched with the Device instances by their name and
// provide the read/write functionality for Devices. If a DeviceHandler is not
// registered for a Device, the Device will not be usable by the plugin.
func (plugin *Plugin) RegisterDeviceHandlers(handlers ...*DeviceHandler) error {
	return plugin.deviceManager.AddHandlers(handlers...)
}

// RegisterDeviceSetupActions registers actions with the device manager which will be
// executed on start. These actions are used for device-specific setup.
//
// fixme: no more kind, need to fix the below.
//
// The filter parameter should be the filter to apply to devices. Currently
// filtering is supported for device kind and type. Filter strings are specified in
// the format "key=value,key=value". The filter
//     "kind=temperature,kind=ABC123"
// would only match devices whose kind was temperature or ABC123.
func (plugin *Plugin) RegisterDeviceSetupActions(filter string, actions ...*DeviceAction) {
	plugin.deviceManager.AddDeviceSetupActions(filter, actions...)
}

// initialize initializes the plugin and all plugin components.
func (plugin *Plugin) initialize() error {
	// Plugin setup. This ensures that the plugin is set up and all
	// plugin metadata that we need is present.
	if err := plugin.init(); err != nil {
		return err
	}

	// Initialize all plugin components
	if err := plugin.deviceManager.init(); err != nil {
		return err
	}
	if err := plugin.server.init(); err != nil {
		return err
	}
	return nil
}

// init initializes the plugin.
func (plugin *Plugin) init() error {
	// Ensure command line flags have been parsed.
	flag.Parse()

	// Check if the plugin should run in debug mode.
	if flagDebug || plugin.config.Debug {
		log.SetLevel(log.DebugLevel)
	}

	// The plugin needs a name in order to run.
	if plugin.info.Name == "" {
		// fixme
		return fmt.Errorf("plugin needs a name to run")
	}

	return nil
}

// run runs the plugin by starting all of the configured plugin components.
func (plugin *Plugin) run() error {
	// Start the plugin components. Order matters here.
	plugin.stateManager.Start()
	plugin.scheduler.Start()

	// Run the gRPC server. This will block while running until the
	// plugin is terminated.
	return plugin.server.start()
}

// onQuitSignal is a function that runs as a goroutine during plugin Run. It
// listens for a quit signal and will terminate the plugin when such a signal
// is received.
//
// Post-run actions are executed here as part of plugin termination.
func (plugin *Plugin) onQuitSignal() {
	// Register system calls for graceful stopping.
	signal.Notify(plugin.quit, syscall.SIGTERM)
	signal.Notify(plugin.quit, syscall.SIGINT)

	log.Info("[plugin] will terminate on: [SIGTERM, SIGINT]")

	// Listen for the quit signal(s). This will block until a signal
	// is received.
	sig := <-plugin.quit

	// If we get here, a signal was received, so we can run termination actions.
	log.WithFields(log.Fields{
		"signal": sig.String(),
	}).Info("[plugin] terminating plugin...")

	if err := plugin.execPostRun(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("[plugin] failed post-run action execution")
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

// handleRunOptions checks whether any command line options were specified for
// the plugin run. If any are set, it handles them appropriately.
func (plugin *Plugin) handleRunOptions() {
	var terminate bool

	// --info was set; print the plugin metadata.
	if flagInfo {
		fmt.Println(plugin.info.format())
		terminate = true
	}

	// --version was set; print the plugin version.
	if flagVersion {
		fmt.Println(plugin.version.format())
		terminate = true
	}

	if terminate {
		// fixme: for testing, should we use an Exiter interface?
		os.Exit(0)
	}
}

// loadPluginConfig loads plugin configurations from file and environment
// and marshals that data into the provided Plugin config struct.
func loadPluginConfig(conf *config.Plugin) error {
	// Setup the config loader for the plugin.
	loader := config.NewYamlLoader("plugin")
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
