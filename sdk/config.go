package sdk

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/policies"
)

// Config holds the configuration for a plugin, its device configs, and
// its type configs.
var Config = config{}

// TODO (etd): The organization of all these config pieces are... confusing and
//  not necessarily intuitive. Additionally, we are accessing all of our config
//  globally which isn't technically ideal since any plugin could be free to
//  override it. What we *should* do is load the config and pass it to the thing
//  that needs it. e.g. on init, load the plugin config. The Plugin can take it
//  as a param.
//  There are questions around this too.. like how do we make that info accessible
//  to other components. Also, we have 3 different kinds of configs and they can
//  be referenced from a number of places, so there are a lot of moving parts (hence
//  why this stuff is laid out in a fairly confusing way).

// config is a struct that holds all of the configs.
type config struct {
	Device     *DeviceConfig
	Plugin     *PluginConfig
	OutputType *OutputType
}

// reset clears the config struct.
func (config *config) reset() {
	config.Device = nil
	config.Plugin = nil
	config.OutputType = nil
}

// ConfigComponent is an interface that all structs that define configuration
// components should implement.
//
// This interface implements a Validate function which is used by the
// schemeValidator in order to validate each struct that makes up a configuration.
type ConfigComponent interface {
	Validate(*errors.MultiError)
}

// VersionedConfig is an interface that SDK config structs implement. It allows
// them to be referenced as a class of configuration which has a version attributed
// to them.
//
// The config version is a simple integer. A particular version of the SDK will
// only support their corresponding version. This allows us to do version checks
// as a sort of pre-flight check for the plugin.
type VersionedConfig interface {
	GetVersion() int
}

// ConfigContext is a structure that associates context with configuration info.
//
// The context around some bit of configuration is useful in logging/errors, as
// it lets us know which config we are talking about.
type ConfigContext struct {
	// Source is where the config came from.
	Source string

	// Config is the configuration itself.
	Config VersionedConfig
}

// NewConfigContext creates a new Context instance.
func NewConfigContext(source string, config VersionedConfig) *ConfigContext {
	return &ConfigContext{
		Source: source,
		Config: config,
	}
}

// IsDeviceConfig checks whether the config in this context is a DeviceConfig.
func (ctx *ConfigContext) IsDeviceConfig() bool {
	_, ok := ctx.Config.(*DeviceConfig)
	return ok
}

// IsPluginConfig checks whether the config in the context is a PluginConfig.
func (ctx *ConfigContext) IsPluginConfig() bool {
	_, ok := ctx.Config.(*PluginConfig)
	return ok
}

// IsOutputTypeConfig checks whether the config in the context is an OutputType config.
func (ctx *ConfigContext) IsOutputTypeConfig() bool {
	_, ok := ctx.Config.(*OutputType)
	return ok
}

// processDeviceConfigs searches for, reads, and validates the device configuration(s).
// Its behavior will vary depending on the device config policies that are set. If
// device config is processed successfully, it will be set to the global Device variable.
func processDeviceConfigs() error { // nolint: gocyclo
	// Get the plugin's policy for device config files.
	deviceFilePolicy := policies.GetDeviceConfigFilePolicy()

	// Get the plugin's policy for dynamic device config.
	deviceDynamicPolicy := policies.GetDeviceConfigDynamicPolicy()

	var deviceCtxs []*ConfigContext

	// Now, try getting the device config(s) from file.
	fileCtxs, err := getDeviceConfigsFromFile()

	// If the error is not a "config not found" error, then we will return it.
	if err != nil {
		_, notFoundErr := err.(*errors.ConfigsNotFound)
		if !notFoundErr {
			return err
		}
	}

	// Regardless of whether we pass policy checks/config validation,
	// we will want to see what the configs are, if in debug mode.
	if log.GetLevel() == log.DebugLevel {
		for i, ctx := range fileCtxs {
			cfg := ctx.Config.(*DeviceConfig)
			json, e := cfg.JSON()
			if e != nil {
				log.Errorf("[sdk] failed to marshal device config to json: %v", err)
			} else {
				log.Debugf("[sdk] device config from file [%d]: %v", i, json)
			}
		}
	}

	switch deviceFilePolicy {
	case policies.DeviceConfigFileRequired:
		if err != nil {
			return errors.NewPolicyViolationError(
				deviceFilePolicy.String(),
				fmt.Sprintf("device config file(s) required, but not found: %v", err),
			)
		}

	case policies.DeviceConfigFileOptional:
		if err != nil {
			fileCtxs = []*ConfigContext{}
			log.Debug("[sdk] no device configuration config files found")
		}

	case policies.DeviceConfigFileProhibited:
		// If the device config file is prohibited, we will log a warning
		// if a file is found, but we will ultimately not fail. Instead, we
		// will just pass along an empty config.
		if err == nil && len(fileCtxs) > 0 {
			log.Warn(
				"[sdk] device config file(s) found, but its use is prohibited via policy. " +
					"the device config files will be ignored.",
			)
		}
		fileCtxs = []*ConfigContext{}

	default:
		return errors.NewPolicyViolationError(
			deviceFilePolicy.String(),
			"unsupported device config file policy",
		)
	}
	log.WithField("policy", deviceFilePolicy.String()).Debug("[sdk] policy validation successful")

	// Now, we can append whatever config contexts we got from file to the slice of all
	// device config contexts.
	deviceCtxs = append(deviceCtxs, fileCtxs...)

	var dynamicCtxs []*ConfigContext

	// Get device configs from dynamic registration
	multiErr := errors.NewMultiError("dynamic device config registration")
	for _, dynamicData := range Config.Plugin.DynamicRegistration.Config {
		dynamicCfgs, e := ctx.dynamicDeviceConfigRegistrar(dynamicData)
		if e != nil {
			multiErr.Add(e)
			continue
		}
		for _, cfg := range dynamicCfgs {
			dynamicCtxs = append(dynamicCtxs, NewConfigContext("dynamic registration", cfg))
		}
	}

	// If any of the errors is not a "config not found" error, then we will return it.
	if multiErr.HasErrors() {
		for _, err := range multiErr.Errors {
			_, notFoundErr := err.(*errors.ConfigsNotFound)
			if !notFoundErr {
				return multiErr
			}
		}
	}

	// Regardless of whether we pass policy checks/config validation,
	// we will want to see what the configs are, if in debug mode.
	if log.GetLevel() == log.DebugLevel {
		for i, ctx := range dynamicCtxs {
			cfg := ctx.Config.(*DeviceConfig)
			json, e := cfg.JSON()
			if e != nil {
				log.Errorf("[sdk] failed to marshal device config to json: %v", err)
			} else {
				log.Debugf("[sdk] device config from dynamic registration [%d]: %v", i, json)
			}
		}
	}

	switch deviceDynamicPolicy {
	case policies.DeviceConfigDynamicRequired:
		if multiErr.Err() != nil || len(dynamicCtxs) == 0 {
			return errors.NewPolicyViolationError(
				deviceDynamicPolicy.String(),
				fmt.Sprintf("dynamic device config(s) required, but none found: %v", multiErr),
			)
		}

	case policies.DeviceConfigDynamicOptional:
		if multiErr.Err() != nil {
			dynamicCtxs = []*ConfigContext{}
			log.Debug("[sdk] no dynamic device configuration(s) found")
		}

	case policies.DeviceConfigDynamicProhibited:
		// If dynamic device configs are prohibited, we will log a warning
		// if any are found, but we will ultimately not fail. Instead, we
		// will just pass along an empty config.
		if multiErr.Err() == nil && len(dynamicCtxs) > 0 {
			log.Warn(
				"[sdk] dynamic device config(s) found, but its use is prohibited via policy. " +
					"the device config(s) will be ignored.",
			)
		}
		dynamicCtxs = []*ConfigContext{}

	default:
		return errors.NewPolicyViolationError(
			deviceDynamicPolicy.String(),
			"unsupported dynamic device config policy",
		)
	}
	log.WithField("policy", deviceDynamicPolicy.String()).Debug("[sdk] policy validation successful")

	// Now, we can append whatever config contexts we got from dynamic registration to the slice
	// of all device config contexts.
	deviceCtxs = append(deviceCtxs, dynamicCtxs...)

	// Validate the device configs
	for _, ctx := range deviceCtxs {
		if ctx.Config.GetVersion() != currentDeviceSchemeVersion {
			return errors.NewValidationError(
				ctx.Source,
				"[sdk] the specified configuration version is not supported with this "+
					"version of the SDK.",
			)
		}
	}

	// Unify the device configs. If there are no device configs
	// at this point, we'll just create an empty one.
	var unifiedCtx *ConfigContext
	if len(deviceCtxs) == 0 {
		unifiedCtx = NewConfigContext("empty", &DeviceConfig{
			Version: currentDeviceSchemeVersion,
		})
	} else {
		unifiedCtx, err = unifyDeviceConfigs(deviceCtxs)
		if err != nil {
			return err
		}
	}

	// Verify that the data defined in the configs is correct, references resolve, etc.
	cfg := unifiedCtx.Config.(*DeviceConfig)
	multiErr = verifyConfigs(cfg)
	if multiErr.HasErrors() {
		return multiErr
	}

	// Validate that the `Data` fields in the config are correct using the plugin-specified
	// validator, since `Data` is plugin-specific.
	multiErr = cfg.ValidateDeviceConfigData(ctx.deviceDataValidator)
	if multiErr.HasErrors() {
		return multiErr
	}

	// With the config validated and unified, we can now assign it to the global Device variable.
	Config.Device = cfg
	return nil
}

// processPluginConfig searches for, reads, and validates the plugin configuration.
// Its behavior will vary depending on the plugin config policy that is set. If
// plugin config is processed successfully, it will be set to the global Plugin
// variable.
func processPluginConfig() error { // nolint: gocyclo
	// Get the plugin's policy for plugin config files.
	pluginFilePolicy := policies.GetPluginConfigFilePolicy()

	// Now, try getting the plugin config from file.
	pluginCtx, err := getPluginConfigFromFile()

	// If the error is not a "config not found" error, then we will return it.
	if err != nil {
		_, notFoundErr := err.(*errors.ConfigsNotFound)
		if !notFoundErr {
			return err
		}
	}

	// Regardless of whether we pass policy checks/config validation,
	// we will want to see what the config is. The config can be nil
	// if only defaults are being used.
	if pluginCtx == nil {
		log.Info("[sdk] no config found from file, checking policy and using defaults")
	} else {
		cfg := pluginCtx.Config.(*PluginConfig)
		json, e := cfg.JSON()
		if e != nil {
			log.Errorf("[sdk] failed to marshal plugin config to json: %v", err)
		} else {
			log.Infof("[sdk] plugin config: %v", json)
		}
	}

	switch pluginFilePolicy {
	case policies.PluginConfigFileRequired:
		if err != nil {
			return errors.NewPolicyViolationError(
				pluginFilePolicy.String(),
				fmt.Sprintf("plugin config file required, but not found: %v", err),
			)
		}

	case policies.PluginConfigFileOptional:
		if err != nil {
			ctx, e := NewDefaultPluginConfig()
			if e != nil {
				return e
			}
			pluginCtx = NewConfigContext("default", ctx)
		}

	case policies.PluginConfigFileProhibited:
		// If the plugin config file is prohibited, we will log a warning
		// if a file is found, but we will ultimately not fail. Instead, we
		// will just pass along an empty config.
		//
		// It is up to the user to specify the config (whether default of not)
		// when the plugin config is prohibited.
		if err == nil && pluginCtx != nil {
			log.Warn(
				"[sdk] plugin config file found, but its use is prohibited via policy. " +
					"you must ensure that the plugin has its config set manually.",
			)
		}
		// The user should have specified the config, so we will take
		// that config and wrap it in a context for validation.
		if Config.Plugin == nil {
			return errors.NewPolicyViolationError(
				pluginFilePolicy.String(),
				"plugin config prohibited via file and not set manually",
			)
		}
		pluginCtx = NewConfigContext("user defined", Config.Plugin)

	default:
		return errors.NewPolicyViolationError(
			pluginFilePolicy.String(),
			"unsupported plugin config file policy",
		)
	}
	log.WithField("policy", pluginFilePolicy.String()).Debug("[sdk] policy validation successful")

	// Validate the plugin config
	if pluginCtx.Config.GetVersion() != currentDeviceSchemeVersion {
		return errors.NewValidationError(
			pluginCtx.Source,
			"[sdk] the specified configuration version is not supported with this "+
				"version of the SDK.",
		)
	}

	// With the config validated, we can now assign it to the global Plugin variable.
	Config.Plugin = pluginCtx.Config.(*PluginConfig)
	return nil
}

// processOutputTypeConfig searches for, reads, and validates the output type
// configuration from file. Its behavior will vary depending on the output type
// config policy that is set. If output type config is processed successfully,
// the found output type configs are returned.
func processOutputTypeConfig() ([]*OutputType, error) { // nolint: gocyclo
	// Get the plugin's policy for output type config files.
	outputTypeFilePolicy := policies.GetTypeConfigFilePolicy()

	// Now, try getting the output type config(s) from file.
	outputTypeCtxs, err := getOutputTypeConfigsFromFile()

	// If the error is not a "config not found" error, then we will return it.
	if err != nil {
		_, notFoundErr := err.(*errors.ConfigsNotFound)
		if !notFoundErr {
			return nil, err
		}
	}

	// Regardless of whether we pass policy checks/config validation,
	// we will want to see what the configs are, if in debug mode.
	if log.GetLevel() == log.DebugLevel {
		for i, ctx := range outputTypeCtxs {
			cfg := ctx.Config.(*OutputType)
			json, e := cfg.JSON()
			if e != nil {
				log.Errorf("[sdk] failed to marshal output type config to json: %v", err)
			} else {
				log.Debugf("[sdk] output type config [%d]: %v", i, json)
			}
		}
	}

	switch outputTypeFilePolicy {
	case policies.TypeConfigFileRequired:
		if err != nil {
			return nil, errors.NewPolicyViolationError(
				outputTypeFilePolicy.String(),
				fmt.Sprintf("output type config file(s) required, but not found: %v", err),
			)
		}

	case policies.TypeConfigFileOptional:
		if err != nil {
			outputTypeCtxs = []*ConfigContext{}
			log.Debug("[sdk] no type configuration config files found")
		}

	case policies.TypeConfigFileProhibited:
		// If the output type config file is prohibited, we will log a warning
		// if a file is found, but we will ultimately not fail. Instead, we
		// will just pass along an empty config.
		if err == nil && len(outputTypeCtxs) > 0 {
			log.Warn(
				"[sdk] output type config file(s) found, but its use is prohibited via policy. " +
					"the output type config files will be ignored.",
			)
			outputTypeCtxs = []*ConfigContext{}
		}

	default:
		return nil, errors.NewPolicyViolationError(
			outputTypeFilePolicy.String(),
			"unsupported output type config file policy",
		)
	}
	log.WithField("policy", outputTypeFilePolicy.String()).Debug("[sdk] policy validation successful")

	var outputs []*OutputType

	// Validate the plugin config
	for _, outputTypeCtx := range outputTypeCtxs {
		if outputTypeCtx.Config.GetVersion() != currentDeviceSchemeVersion {
			return nil, errors.NewValidationError(
				outputTypeCtx.Source,
				"[sdk] the specified configuration version is not supported with this "+
					"version of the SDK.",
			)
		}

		cfg := outputTypeCtx.Config.(*OutputType)
		outputs = append(outputs, cfg)
	}
	return outputs, nil
}

// unifyDeviceConfigs will take a slice of ConfigContext which represents
// DeviceConfigs and unify them into a single ConfigContext for a DeviceConfig.
//
// If any of the ConfigContexts given as a parameter do not represent a
// DeviceConfig, an error is returned.
func unifyDeviceConfigs(ctxs []*ConfigContext) (*ConfigContext, error) {

	// FIXME (etd): figure out how to either:
	//  i. merge the source info into the ConfigContext
	// ii. map each component to its original context so we know exactly where
	//     a specific field/config component originated from.

	// If there are no contexts, we can't unify.
	if len(ctxs) == 0 {
		return nil, fmt.Errorf("no ConfigContexts specified for unification")
	}

	log.Debugf("[sdk] unifying %d device configs", len(ctxs))

	var context *ConfigContext
	for _, ctx := range ctxs {
		if !ctx.IsDeviceConfig() {
			return nil, fmt.Errorf("config context does not represent a device config")
		}
		if context == nil {
			context = ctx
		} else {
			base := context.Config.(*DeviceConfig)
			source := ctx.Config.(*DeviceConfig)

			// Merge DeviceConfig.Locations - config verification will ensure that these
			// are unique.
			base.Locations = append(base.Locations, source.Locations...)

			// Merge DeviceConfig.Devices - generally deviceKinds should not be defined in
			// multiple files, but if doing dynamic registration, it likely will come in this
			// way. as a result, we will need to merge instance/output data for device kinds with
			// the same name..
			// FIXME: without checking that the device kinds are actually the same, there
			// there may be some dragons lurking here.
			mergeDeviceKinds(&base.Devices, &source.Devices)
		}
	}
	return context, nil
}

// mergeDeviceKinds will add the device kinds from the `source` into the `base` if
// a device kind with that name does not exist in the base, and will merge the device
// kind fields if it does exist.
func mergeDeviceKinds(base, source *[]*DeviceKind) {
	exists := map[string]*DeviceKind{}
	for _, kind := range *base {
		exists[kind.Name] = kind
	}

	for _, kind := range *source {
		k, found := exists[kind.Name]
		if !found {
			// If it is not found, add it to the base slice
			*base = append(*base, kind)
		} else {
			// Otherwise, just update the kind that is already in the base slice
			k.Instances = append(k.Instances, kind.Instances...)
		}
	}
}
