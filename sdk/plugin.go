// Synse SDK
// Copyright (c) 2019 Vapor IO
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package sdk

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/health"
	"github.com/vapor-ware/synse-sdk/sdk/output"
	"github.com/vapor-ware/synse-sdk/sdk/policy"
)

const (
	// PluginEnvOverride defines the environment variable that can be used to
	// set an override config location for the Plugin configuration file.
	PluginEnvOverride = "PLUGIN_CONFIG"
)

var (
	// Command line arguments
	flagDebug   bool
	flagVersion bool
	flagDryRun  bool

	// Config file locations
	currentDirConfig    = "."
	localPluginConfig   = "./config"
	defaultPluginConfig = "/etc/synse/plugin/config"
)

func init() {
	flag.BoolVar(&flagDebug, "debug", false, "enable debug logging")
	flag.BoolVar(&flagVersion, "version", false, "print the plugin version information")
	flag.BoolVar(&flagDryRun, "dry-run", false, "run only the setup actions to verify functionality and configuration")

	flag.Parse()
	handleRunOptions()
}

// PluginAction defines an action that can be run before or after the main
// Plugin run logic. This is generally used for setup/teardown.
type PluginAction struct {
	Name   string
	Action func(p *Plugin) error
}

// Plugin is a Synse Plugin.
type Plugin struct {
	info     *PluginMetadata
	version  *pluginVersion
	config   *config.Plugin
	quit     chan os.Signal
	id       *pluginID
	policies *policy.Policies

	// Actions
	preRun  []*PluginAction
	postRun []*PluginAction

	// Options and handlers
	pluginHandlers *PluginHandlers

	// Plugin components
	scheduler *scheduler
	state     *stateManager
	device    *deviceManager
	server    *server
	health    *health.Manager
}

// NewPlugin creates a new instance of a Plugin. This should be the only
// way that a Plugin is initialized.
//
// This constructor will load the plugin configuration; if it is not present
// or invalid, this will fail. All other Plugin component initialization
// is deferred until Run is called.
func NewPlugin(options ...PluginOption) (*Plugin, error) {
	// Since this is essentially the entry point for the plugin and setup actions
	// occur as part of plugin construction, we want to set the log level as early
	// as possible. If the debug flag is set, set the level to debug.
	if flagDebug {
		log.SetLevel(log.DebugLevel)
	}

	// Various things use the plugin metadata on setup, so we need to make sure
	// it is set prior to initializing the plugin.
	if metadata.Name == "" {
		return nil, fmt.Errorf(
			"plugin metadata must be set prior to calling 'NewPlugin()'; " +
				"this can be done via 'sdk.SetPluginInfo()'",
		)
	}

	log.Debug("[plugin] creating new plugin")

	// Create the plugin. We create the instance first so a reference to it
	// is available for subsequent setup actions.
	p := Plugin{
		version:        version,
		info:           &metadata,
		config:         new(config.Plugin),
		quit:           make(chan os.Signal),
		policies:       policy.NewDefaultPolicies(),
		pluginHandlers: NewDefaultPluginHandlers(),
	}

	// Set custom options for the plugin.
	log.WithField("options", len(options)).Debug("[plugin] loading plugin options")
	for _, option := range options {
		option(&p)
	}

	// Load the plugin configuration.
	if err := p.loadConfig(); err != nil {
		log.Errorf("[plugin] failed to load plugin config")
		return nil, err
	}

	// Check if debug mode was set in the plugin config. If so, set the log level
	// to debug here.
	if p.config.Debug {
		log.SetLevel(log.DebugLevel)
	}

	// Log the plugin metadata, version info, and config.
	metadata.log()
	version.Log()
	p.config.Log()

	// Initialize the plugin ID namespace.
	id, err := newPluginID(p.config.ID, &metadata)
	if err != nil {
		log.Error("[plugin] failed to initialize plugin ID namespace")
		return nil, err
	}
	p.id = id

	// Initialize the plugin components. The order in which components are initialized
	// is important, since a dependency chain exists between some components. In particular:
	// * the scheduler requires the device manager and state manager
	// * the server requires the device manager, state manager, scheduler, and health manager
	p.health = health.NewManager(p.config.Health)
	p.state = newStateManager(p.config.Settings)
	p.device = newDeviceManager(&p)
	p.scheduler = newScheduler(&p)
	p.server = newServer(&p)

	return &p, nil
}

// Run starts the plugin.
//
// This is the functional starting point for all plugins. Once this is called,
// the plugin will initialize all of its components and validate its state. Once
// everything is ready, it will run each of its components. The gRPC server is
// run in the foreground; all other components are run as goroutines.
func (plugin *Plugin) Run() error {
	// Initialize the plugin and its components.
	if err := plugin.initialize(); err != nil {
		log.Error("[plugin] failed to initialize plugin")
		return err
	}

	// If all components initialized without error, we can register
	// any pre/post run actions which they may have.
	plugin.state.registerActions(plugin)
	plugin.scheduler.registerActions(plugin)
	plugin.server.registerActions(plugin)

	// Run pre-run actions, if any exist.
	if err := plugin.execPreRun(); err != nil {
		log.Error("[plugin] failed to execute plugin pre-run actions")
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

	// Run the plugin.
	return plugin.run()
}

// RegisterHealthChecks registers custom health checks with the plugin.
func (plugin *Plugin) RegisterHealthChecks(checks ...health.Check) error {
	for _, check := range checks {
		if err := plugin.health.Register(check); err != nil {
			return err
		}
	}
	return nil
}

// RegisterOutputs registers new Outputs with the plugin. A plugin will automatically
// register the built-in SDK outputs. This function allows a plugin do augment that
// set of outputs with its own custom outputs.
//
// If any registered output names conflict with those of built-in or other custom
// outputs, an error is returned.
func (plugin *Plugin) RegisterOutputs(outputs ...*output.Output) error {
	return output.Register(outputs...)
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
	return plugin.device.AddHandlers(handlers...)
}

// RegisterDeviceSetupActions registers actions with the device manager which will be
// executed on start. These actions are used for device-specific setup.
func (plugin *Plugin) RegisterDeviceSetupActions(actions ...*DeviceAction) error {
	return plugin.device.AddDeviceSetupActions(actions...)
}

// initialize initializes the plugin and all plugin components.
func (plugin *Plugin) initialize() error {
	log.Info("[plugin] initializing")

	// Initialize all plugin components
	if err := plugin.device.init(); err != nil {
		return err
	}
	if err := plugin.server.init(); err != nil {
		return err
	}
	return nil
}

// run runs the plugin by starting all of the configured plugin components.
func (plugin *Plugin) run() error {
	log.Info("[plugin] running")

	// Start the Prometheus metrics exporter, if metrics are enabled for
	// the plugin. This is a blocking function so it must be called in a goroutine.
	if plugin.config.Metrics.Enabled {
		log.Debug("[plugin] application metrics enabled")
		go exposeMetrics()
	}

	// Start the plugin components. Order matters here.
	if err := plugin.device.Start(plugin); err != nil {
		log.Error("[plugin] failed to start device manager")
		return err
	}
	plugin.state.Start()
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
	}).Info("[plugin] terminating plugin")

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

// loadConfig loads plugin configurations from file and environment
// and marshals that data into the Plugin's config struct.
func (plugin *Plugin) loadConfig() error {
	// Setup the config loader for the plugin.
	loader := config.NewYamlLoader("plugin")
	loader.EnvPrefix = "PLUGIN"
	loader.EnvOverride = PluginEnvOverride
	loader.FileName = "config"
	loader.AddSearchPaths(
		currentDirConfig,
		localPluginConfig,
		defaultPluginConfig,
	)

	// Load the plugin configuration.
	if err := loader.Load(plugin.policies.PluginConfig); err != nil {
		log.WithField("error", err).Error("[plugin] failed to load plugin configuration")
		return err
	}

	// Marshal the configuration into the plugin config struct.
	return loader.Scan(plugin.config)
}

// handleRunOptions checks whether any command line options were specified for
// the plugin run. If any are set, it handles them appropriately.
func handleRunOptions() {
	var terminate bool

	if flagDebug {
		log.SetLevel(log.DebugLevel)
	}

	// --version was set; print the plugin version.
	if flagVersion {
		fmt.Println(version.format())
		terminate = true
	}

	if terminate {
		// fixme: for testing, should we use an Exiter interface?
		os.Exit(0)
	}
}
