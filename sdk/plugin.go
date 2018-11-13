package sdk

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/creasty/defaults"
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

// setup performs the pre-run setup actions for a plugin.
func (plugin *Plugin) setup() error {
	// Register system calls for graceful stopping.
	signal.Notify(plugin.quit, syscall.SIGTERM)
	signal.Notify(plugin.quit, syscall.SIGINT)
	go plugin.onQuit()

	// The plugin name must be set as metainfo, since it is used in the Device
	// model. Check if it is set here. If not, return an error.
	if metainfo.Name == "" {
		return fmt.Errorf("plugin name not set, but required; see sdk.SetPluginMetainfo")
	}

	// Check for command line flags. If any flags are set that require an
	// action, that action will be resolved here.
	parseFlags()

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

// The current (latest) version of the plugin config scheme.
var currentPluginSchemeVersion = "1.0"

// NewDefaultPluginConfig creates a new instance of a PluginConfig with its
// default values resolved.
func NewDefaultPluginConfig() (*PluginConfig, error) {
	config := &PluginConfig{
		SchemeVersion: SchemeVersion{Version: currentPluginSchemeVersion},
	}
	err := defaults.Set(config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// PluginConfig contains the configuration options for the plugin.
type PluginConfig struct {

	// SchemeVersion is the version of the configuration scheme.
	SchemeVersion `yaml:",inline"`

	// Debug is a flag that determines whether the plugin should run
	// with debug logging or not.
	Debug bool `default:"false" yaml:"debug,omitempty" addedIn:"1.0"`

	// Settings provide specifications for how the plugin should run.
	Settings *PluginSettings `default:"{}" yaml:"settings,omitempty" addedIn:"1.0"`

	// Network specifies the networking configuration for the plugin.
	Network *NetworkSettings `default:"{\"type\": \"tcp\", \"address\": \"localhost:5001\"}" yaml:"network,omitempty" addedIn:"1.0"`

	// DynamicRegistration specifies configuration settings and data
	// for how the plugin should handle dynamic device registration.
	DynamicRegistration *DynamicRegistrationSettings `default:"{}" yaml:"dynamicRegistration,omitempty" addedIn:"1.0"`

	// Limiter specifies settings for a rate limiter for reads/writes.
	Limiter *LimiterSettings `yaml:"limiter,omitempty" addedIn:"1.0"`

	// Health specifies the settings for health checking in the plugin.
	Health *HealthSettings `default:"{}" yaml:"health,omitempty" addedIn:"1.0"`

	// Context is a map that allows the plugin to specify any arbitrary
	// data it may need.
	Context map[string]interface{} `default:"{}" yaml:"context,omitempty" addedIn:"1.0"`
}

// JSON encodes the config as JSON. This can be useful for logging and debugging.
func (config *PluginConfig) JSON() (string, error) {
	bytes, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Validate validates that the PluginConfig has no configuration errors.
func (config PluginConfig) Validate(multiErr *errors.MultiError) {
	// A version must be specified and it must be of the correct format.
	_, err := config.GetVersion()
	if err != nil {
		log.WithField("config", config).Error("[validation] bad version")
		multiErr.Add(errors.NewValidationError(multiErr.Context["source"], err.Error()))
	}

	// If network is nil or an empty struct, error. We need to know how
	// the plugin should communicate with Synse server.
	if config.Network == nil || config.Network == (&NetworkSettings{}) {
		log.WithField("config", config).Error("[validation] no network")
		multiErr.Add(errors.NewFieldRequiredError(multiErr.Context["source"], "network"))
	}
}

// PluginSettings specifies the configuration options that determine the
// runtime behavior of the plugin.
type PluginSettings struct {
	// Mode is the run mode of the read and write loops. This can either
	// be "serial" or "parallel".
	Mode string `default:"serial" yaml:"mode,omitempty" addedIn:"1.0"`

	// Listen contains the settings to configure listener behavior.
	Listen *ListenSettings `default:"{}" yaml:"listen,omitempty" addedIn:"1.2"`

	// Read contains the settings to configure read behavior.
	Read *ReadSettings `default:"{}" yaml:"read,omitempty" addedIn:"1.0"`

	// Write contains the settings to configure write behavior.
	Write *WriteSettings `default:"{}" yaml:"write,omitempty" addedIn:"1.0"`

	// Transaction contains the settings to configure transaction
	// handling behavior.
	Transaction *TransactionSettings `default:"{}" yaml:"transaction,omitempty" addedIn:"1.0"`

	// Cache contains the settings to configure local data caching
	// by the plugin.
	Cache *CacheSettings `default:"{}" yaml:"cache,omitempty" addedIn:"1.2"`
}

// Validate validates that the PluginSettings has no configuration errors.
func (settings PluginSettings) Validate(multiErr *errors.MultiError) {
	if settings.Mode != modeSerial && settings.Mode != modeParallel {
		log.WithField("config", settings).Error("[validation] bad mode")
		multiErr.Add(errors.NewInvalidValueError(
			multiErr.Context["source"],
			"settings.mode",
			"one of: serial, parallel",
		))
	}
}

// IsSerial checks if the PluginSettings is configured with mode "serial".
func (settings *PluginSettings) IsSerial() bool {
	return settings.Mode == modeSerial
}

// IsParallel checks if the PluginSettings is configured with mode "parallel".
func (settings *PluginSettings) IsParallel() bool {
	return settings.Mode == modeParallel
}

// NetworkSettings specifies the configuration options around the gRPC
// server's networking behavior.
type NetworkSettings struct {
	// Type is the type of networking. Currently, this must be one of
	// "tcp" (TCP/IP) or "unix" (Unix Socket)
	Type string `yaml:"type,omitempty" addedIn:"1.0"`

	// Address is the address to communicate over. For "tcp", this would
	// be the host/port (e.g. 0.0.0.0:50001). For "unix", this would be
	// the name of the unix socket (e.g. plugin.sock).
	Address string `yaml:"address,omitempty" addedIn:"1.0"`

	// TLS contains the configuration settings for TLS/SSL for the gRPC
	// connection between Synse Server and the plugin. If this is not set,
	// insecure transport will be used.
	TLS *TLSNetworkSettings `yaml:"tls,omitempty" addedIn:"1.1"`
}

// Validate validates that the NetworkSettings has no configuration errors.
func (settings NetworkSettings) Validate(multiErr *errors.MultiError) {
	if settings.Type == "" {
		log.WithField("config", settings).Error("[validation] empty type")
		multiErr.Add(errors.NewFieldRequiredError(multiErr.Context["source"], "network.type"))
	} else {
		if settings.Type != networkTypeTCP && settings.Type != networkTypeUnix {
			log.WithField("config", settings).Error("[validation] bad type")
			multiErr.Add(errors.NewInvalidValueError(
				multiErr.Context["source"],
				"network.type",
				"one of: unix, tcp",
			))
		}
	}
	if settings.Address == "" {
		log.WithField("config", settings).Error("[validation] empty address")
		multiErr.Add(errors.NewFieldRequiredError(multiErr.Context["source"], "network.address"))
	}
}

// TLSNetworkSettings specifies configuration around TLS/SSL for securing the
// gRPC communication layer between Synse Server and plugins using this SDK.
type TLSNetworkSettings struct {
	// Cert is the location of the cert file to use for the gRPC server.
	Cert string `yaml:"cert,omitempty" addedIn:"1.1"`

	// Key is the location of the cert file to use for the gRPC server.
	Key string `yaml:"key,omitempty" addedIn:"1.1"`

	// CACerts are a list of certificate authority certs to use. If none
	// are specified, the OS system-wide TLS certs are used.
	CACerts []string `yaml:"caCerts,omitempty" addedIn:"1.1"`

	// SkipVerify is a flag that, when set, will skip certificate checks.
	SkipVerify bool `yaml:"skipVerify,omitempty" addedIn:"1.1"`
}

// DynamicRegistrationSettings specifies configuration and data for
// the dynamic registration of devices.
type DynamicRegistrationSettings struct {
	// The plugin configuration for dynamic registration. This slice of maps holds the
	// plugin-specific data that can be used to dynamically register new devices.
	// As an example, this could hold the information for connecting with a server,
	// or it could contain a bus address, etc.
	Config []map[string]interface{} `default:"[]" yaml:"config,omitempty" addedIn:"1.0"`
}

// Validate validates that the DynamicRegistrationSettings has no configuration errors.
func (settings DynamicRegistrationSettings) Validate(multiErr *errors.MultiError) {
	// nothing to validate here.
}

// LimiterSettings specifies configurations for a rate limiter on reads
// and writes.
type LimiterSettings struct {
	// Rate is the limit, or maximum frequency of events. A rate of
	// 0 signifies 'unlimited'.
	Rate int `yaml:"rate,omitempty" addedIn:"1.0"`

	// Burst defines the bucket size for the limiter, or maximum number
	// of events that can be fulfilled at once. If this is 0, it will take
	// the same value as the rate.
	Burst int `yaml:"burst,omitempty" addedIn:"1.0"`
}

// Validate validates that the LimiterSettings has no configuration errors.
func (settings LimiterSettings) Validate(multiErr *errors.MultiError) {
	if settings.Rate < 0 {
		log.WithField("config", settings).Error("[validation] bad rate")
		multiErr.Add(errors.NewInvalidValueError(
			multiErr.Context["source"],
			"limiter.rate",
			"greater than or equal to 0",
		))
	}

	if settings.Burst < 0 {
		log.WithField("config", settings).Error("[validation] bad burst")
		multiErr.Add(errors.NewInvalidValueError(
			multiErr.Context["source"],
			"limiter.burst",
			"greater than or equal to 0",
		))
	}
}

// ListenSettings provides configuration options for listener operations.
// A listener is a function that is used to collect push-based data.
type ListenSettings struct {
	// Enabled globally enables or disables listening for the plugin.
	// By default a plugin will have listening enabled.
	Enabled bool `default:"true" yaml:"enabled,omitempty" addedIn:"1.2"`

	// Buffer defines the size of the listen buffer. This will be the
	// size of the channel that passes all the collected push data from
	// all listener instances to the data manager.
	Buffer int `default:"100" yaml:"buffer,omitempty" addedIn:"1.2"`
}

// Validate validates that the ListenSettings has no confiugration errors.
func (settings ListenSettings) Validate(multiErr *errors.MultiError) {
	// If the buffer size is set to 0, return an error. A size
	// of 0 would prevent any data from being moved around, blocking
	// all listen operations.
	if settings.Buffer <= 0 {
		log.WithField("config", settings).Error("[validation] bad listen buffer")
		multiErr.Add(errors.NewInvalidValueError(
			multiErr.Context["source"],
			"settings.listen.buffer",
			"a value greater than 0",
		))
	}
}

// ReadSettings provides configuration options for read operations.
type ReadSettings struct {
	// Enabled globally enables or disables reading for the plugin.
	// By default, a plugin will have reading enabled.
	Enabled bool `default:"true" yaml:"enabled,omitempty" addedIn:"1.0"`

	// Interval specifies the interval at which devices should be
	// read from. This is 1s by default.
	Interval string `default:"1s" yaml:"interval,omitempty" addedIn:"1.0"`

	// Buffer defines the size of the read buffer. This will be
	// the size of the channel that passes along read responses.
	Buffer int `default:"100" yaml:"buffer,omitempty" addedIn:"1.0"`

	// SerialReadInterval specifies the interval to pause between serial reads.
	// This is here to avoid overwhelming a device. This is 0s by default.
	SerialReadInterval string `default:"0s" yaml:"serialReadInterval,omitempty" addedIn:"1.3"`
}

// Validate validates that the ReadSettings has no configuration errors.
func (settings ReadSettings) Validate(multiErr *errors.MultiError) {
	// Try parsing the interval to validate it is a correctly specified duration string.
	_, err := settings.GetInterval()
	if err != nil {
		log.WithField("config", settings).Error("[validation] bad interval")
		multiErr.Add(errors.NewValidationError(multiErr.Context["source"], err.Error()))
	}

	// Try parsing the serial read interval to validate it is a correctly specified duration string.
	_, err = settings.GetSerialReadInterval()
	if err != nil {
		log.WithField("config", settings).Error("[validation] bad serial read interval")
		multiErr.Add(errors.NewValidationError(multiErr.Context["source"], err.Error()))
	}

	// If the buffer size is set to 0, return an error. Previously, this
	// was allowed, as a size of 0 could indicate "no read", but now we
	// have the 'enabled' field, so we don't need to support this.
	if settings.Buffer <= 0 {
		log.WithField("config", settings).Error("[validation] bad read buffer")
		multiErr.Add(errors.NewInvalidValueError(
			multiErr.Context["source"],
			"settings.read.buffer",
			"a value greater than 0",
		))
	}
}

// GetInterval gets the read interval as a duration. If the config
// has been validated successfully, this should never return an error.
func (settings *ReadSettings) GetInterval() (time.Duration, error) {
	return time.ParseDuration(settings.Interval)
}

// GetSerialReadInterval gets the duration to wait between serial reads.
// This is here to avoid overwhelming a device.
func (settings *ReadSettings) GetSerialReadInterval() (time.Duration, error) {
	log.Infof("GetSerialReadInterval settings: %v", settings)
	return time.ParseDuration(settings.SerialReadInterval)
}

// WriteSettings provides configuration options for write operations.
type WriteSettings struct {
	// Enabled globally enables or disables writing for the plugin.
	// By default, a plugin will have writing enabled.
	Enabled bool `default:"true" yaml:"enabled,omitempty" addedIn:"1.0"`

	// Interval specifies the interval at which devices should be
	// written to. This is 1s by default.
	Interval string `default:"1s" yaml:"interval,omitempty" addedIn:"1.0"`

	// Buffer defines the size of the write buffer. This will be
	// the size of the channel that passes along write requests.
	Buffer int `default:"100" yaml:"buffer,omitempty" addedIn:"1.0"`

	// Max is the maximum number of write transactions to process
	// in a single batch. In general, this can tune performance when
	// running in serial mode.
	Max int `default:"100" yaml:"max,omitempty" addedIn:"1.0"`
}

// Validate validates that the WriteSettings has no configuration errors.
func (settings WriteSettings) Validate(multiErr *errors.MultiError) {
	// Try parsing the interval to validate it is a correctly specified duration string.
	_, err := settings.GetInterval()
	if err != nil {
		log.WithField("config", settings).Error("[validation] bad interval")
		multiErr.Add(errors.NewValidationError(multiErr.Context["source"], err.Error()))
	}

	// If the buffer size is set to 0, return an error. Previously, this
	// was allowed, as a size of 0 could indicate "no write", but now we
	// have the 'enabled' field, so we don't need to support this.
	if settings.Buffer <= 0 {
		log.WithField("config", settings).Error("[validation] bad write buffer")
		multiErr.Add(errors.NewInvalidValueError(
			multiErr.Context["source"],
			"settings.write.buffer",
			"a value greater than 0",
		))
	}

	if settings.Max <= 0 {
		log.WithField("config", settings).Error("[validation] bad write max")
		multiErr.Add(errors.NewInvalidValueError(
			multiErr.Context["source"],
			"settings.write.max",
			"a value greater than 0",
		))
	}
}

// GetInterval gets the write interval as a duration. If the config
// has been validated successfully, this should never return an error.
func (settings WriteSettings) GetInterval() (time.Duration, error) {
	return time.ParseDuration(settings.Interval)
}

// TransactionSettings provides configuration options for transaction operations.
type TransactionSettings struct {
	// TTL is the time-to-live for a transaction in the transaction cache.
	TTL string `default:"5m" yaml:"ttl,omitempty" addedIn:"1.0"`
}

// Validate validates that the TransactionSettings has no configuration errors.
func (settings TransactionSettings) Validate(multiErr *errors.MultiError) {
	// Try parsing the interval to validate it is a correctly specified duration string.
	_, err := settings.GetTTL()
	if err != nil {
		log.WithField("config", settings).Error("[validation] bad ttl")
		multiErr.Add(errors.NewValidationError(multiErr.Context["source"], err.Error()))
	}
}

// GetTTL gets the transaction TTL as a duration. If the config has been
// validated successfully, this should never return an error.
func (settings *TransactionSettings) GetTTL() (time.Duration, error) {
	return time.ParseDuration(settings.TTL)
}

// HealthSettings provides configuration options around health checking in
// the plugin.
type HealthSettings struct {
	// UseDefaults determines whether the plugin should use the built-in health
	// checks or not.
	UseDefaults bool `default:"true" yaml:"useDefaults,omitempty" addedIn:"1.0"`
}

// Validate validates that the HealthSettings has no configuration errors.
func (settings HealthSettings) Validate(multiErr *errors.MultiError) {
	// Nothing to validate
}

// CacheSettings provides configuration options for an in-memory windowed
// cache for plugin readings.
type CacheSettings struct {
	// Enabled sets whether the plugin will use a local
	// in-memory cache to store a small window of readings.
	// By default, the cache is not enabled.
	Enabled bool `default:"false" yaml:"enabled,omitempty" addedIn:"1.2"`

	// TTL is the time-to-live for a reading in the readings cache.
	// This will only be used if the cache is enabled. Once a reading
	// has exceeded its TTL, it will be removed from the cache.
	TTL time.Duration `default:"3m" yaml:"ttl,omitempty" addedIn:"1.2"`
}

// Validate validates that the CacheSettings has no configuration errors.
func (settings CacheSettings) Validate(multiErr *errors.MultiError) {
	// Nothing to validate
}
