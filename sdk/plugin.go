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

	"github.com/vapor-ware/synse-sdk/sdk/policy"

	log "github.com/Sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/output"
)

const (
	// PluginEnvOverride defines the environment variable that can be used to
	// set an override config location for the Plugin configuration file.
	PluginEnvOverride = "PLUGIN_CONFIG"
)

var (
	flagDebug   bool
	flagVersion bool
	flagDryRun  bool
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

	// Plugin outputs
	outputs map[string]*output.Output

	// Options and handlers
	pluginHandlers *PluginHandlers

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
	if metadata.Name == "" {
		return nil, fmt.Errorf(
			"plugin metadata must be set prior to calling 'NewPlugin()'; " +
				"this can be done via 'sdk.SetPluginInfo()'",
		)
	}

	// Load the plugin configuration.
	conf := new(config.Plugin)
	// FIXME: we should not hardcode the policy here.. we should see whether any plugin
	//  options specify a plugin policy first. keeping this as-is for now, but we will
	//  need to change this. this will likely require a change to how the plugin is
	//  constructed entirely here.
	if err := loadPluginConfig(conf, policy.Optional); err != nil {
		return nil, err
	}

	// Check if the plugin should run in debug mode.
	if flagDebug || conf.Debug {
		log.SetLevel(log.DebugLevel)
	}

	// Initialize the plugin ID namespace.
	id, err := newPluginID(conf.ID, &metadata)
	if err != nil {
		return nil, err
	}

	pluginHandlers := NewDefaultPluginHandlers()
	pluginPolicies := policy.NewDefaultPolicies()

	// Initialize plugin components.
	dm := newDeviceManager(id, pluginHandlers, pluginPolicies)
	sm := NewStateManager(conf.Settings)
	sched := NewScheduler(conf.Settings, dm, sm)
	server := newServer(conf.Network, dm, sm, sched, &metadata)

	p := Plugin{
		outputs:        make(map[string]*output.Output),
		quit:           make(chan os.Signal),
		info:           &metadata,
		version:        version,
		config:         conf,
		id:             id,
		policies:       pluginPolicies,
		pluginHandlers: pluginHandlers,
		deviceManager:  dm,
		stateManager:   sm,
		scheduler:      sched,
		server:         server,
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

// Run starts the plugin.
//
// This is the functional starting point for all plugins. Once this is called,
// the plugin will initialize all of its components and validate its state. Once
// everything is ready, it will run each of its components. The gRPC server is
// run in the foreground; all other components are run as goroutines.
func (plugin *Plugin) Run() error {
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
func (plugin *Plugin) RegisterDeviceSetupActions(actions ...*DeviceAction) error {
	return plugin.deviceManager.AddDeviceSetupActions(actions...)
}

// initialize initializes the plugin and all plugin components.
func (plugin *Plugin) initialize() error {
	// Initialize all plugin components
	if err := plugin.deviceManager.init(); err != nil {
		return err
	}
	if err := plugin.server.init(); err != nil {
		return err
	}
	return nil
}

// run runs the plugin by starting all of the configured plugin components.
func (plugin *Plugin) run() error {
	// Start the plugin components. Order matters here.
	if err := plugin.deviceManager.Start(plugin); err != nil {
		return err
	}
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

// loadPluginConfig loads plugin configurations from file and environment
// and marshals that data into the provided Plugin config struct.
func loadPluginConfig(conf *config.Plugin, pol policy.Policy) error {
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
	if err := loader.Load(pol); err != nil {
		return err
	}

	// Marshal the configuration into the plugin config struct.
	return loader.Scan(conf)
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
