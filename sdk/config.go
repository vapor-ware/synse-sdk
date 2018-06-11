package sdk

import (
	"fmt"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
	"github.com/vapor-ware/synse-sdk/sdk/policies"
	"strings"
	"strconv"
)


// ConfigComponent is an interface that all structs that define configuration
// components should implement.
//
// This interface implements a Validate function which is used by the
// SchemeValidator in order to validate each struct that makes up a configuration.
type ConfigComponent interface {
	Validate(*errors.MultiError)
}

// ConfigBase is an interface that the base configuration struct should
// implement. This allows the SchemeValidator to get the SchemeVersion
// for that given configuration.
type ConfigBase interface {
	GetVersion() (*ConfigVersion, error)
}

// ConfigContext is a structure that associates context with configuration info.
//
// The context around some bit of configuration is useful in logging/errors, as
// it lets us know which config we are talking about.
type ConfigContext struct {
	// Source is where the config came from.
	Source string

	// Config is the configuration itself. This should be a configuration struct
	// that implements Base. That is to say, the config held in this context
	// should be the root config struct for that config type. This will allow us
	// to get the scheme version of the configuration.
	Config ConfigBase
}

// NewConfigContext creates a new Context instance.
func NewConfigContext(source string, config ConfigBase) *ConfigContext {
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



const (
	tagAddedIn      = "addedIn"
	tagDeprecatedIn = "deprecatedIn"
	tagRemovedIn    = "removedIn"
)

// ConfigVersion is a representation of a configuration scheme version
// that can be compared to other SchemeVersions.
type ConfigVersion struct {
	Major int
	Minor int
}

// NewVersion creates a new instance of a Version.
func NewVersion(versionString string) (*ConfigVersion, error) {
	var min, maj int
	var err error

	if versionString == "" {
		return nil, fmt.Errorf("no version info found")
	}

	s := strings.Split(versionString, ".")
	switch len(s) {
	case 1:
		maj, err = strconv.Atoi(s[0])
		if err != nil {
			return nil, err
		}
		min = 0
	case 2:
		maj, err = strconv.Atoi(s[0])
		if err != nil {
			return nil, err
		}
		min, err = strconv.Atoi(s[1])
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("too many version components - should only have MAJOR[.MINOR]")
	}

	return &ConfigVersion{
		Major: maj,
		Minor: min,
	}, nil
}

// String returns a string representation of the scheme version.
func (version *ConfigVersion) String() string {
	return fmt.Sprintf("%d.%d", version.Major, version.Minor)
}

// IsLessThan returns true if the Version is less than the Version
// provided as a parameter.
func (version *ConfigVersion) IsLessThan(other *ConfigVersion) bool {
	if version.Major < other.Major {
		return true
	}
	if version.Major == other.Major && version.Minor < other.Minor {
		return true
	}
	return false
}

// IsGreaterOrEqualTo returns true if the ConfigVersion is greater than or equal to
// the Version provided as a parameter.
func (version *ConfigVersion) IsGreaterOrEqualTo(other *ConfigVersion) bool {
	if version.Major > other.Major {
		return true
	}
	if version.Major == other.Major && version.Minor >= other.Minor {
		return true
	}
	return false
}

// IsEqual returns true if the Version is equal to the Version provided
// as a parameter.
func (version *ConfigVersion) IsEqual(other *ConfigVersion) bool {
	return version.Major == other.Major && version.Minor == other.Minor
}

// SchemeVersion is a struct that is used to extract the configuration
// scheme version from any config file.
type SchemeVersion struct {
	// Version is the config version scheme specified in the config file.
	Version string `yaml:"version,omitempty" addedIn:"1.0"`

	// scheme is the Version that represents the SchemeVersion's Version.
	scheme *ConfigVersion
}

// parse parses the Version field into a Version.
func (schemeVersion *SchemeVersion) parse() error {
	scheme, err := NewVersion(schemeVersion.Version)
	if err != nil {
		return err
	}
	schemeVersion.scheme = scheme
	return nil
}

// GetVersion gets the Version associated with the version specified
// in the configuration.
func (schemeVersion *SchemeVersion) GetVersion() (*ConfigVersion, error) {
	if schemeVersion.scheme == nil {
		err := schemeVersion.parse()
		if err != nil {
			return nil, err
		}
	}
	return schemeVersion.scheme, nil
}


// processDeviceConfigs searches for, reads, and validates the device configuration(s).
// Its behavior will vary depending on the device config policies that are set. If
// device config is processed successfully, it will be set to the global Device variable.
func processDeviceConfigs() error { // nolint: gocyclo
	// Get the plugin's policy for device config files.
	deviceFilePolicy := policies.GetDeviceConfigFilePolicy()
	logger.Debugf("device config file policy: %s", deviceFilePolicy.String())

	// Get the plugin's policy for dynamic device config.
	deviceDynamicPolicy := policies.GetDeviceConfigDynamicPolicy()
	logger.Debugf("device dynamic config policy: %s", deviceDynamicPolicy.String())

	var deviceCtxs []*config.Context

	// Now, try getting the device config(s) from file.
	fileCtxs, err := config.GetDeviceConfigsFromFile()

	// If the error is not a "config not found" error, then we will return it.
	_, notFoundErr := err.(*errors.ConfigsNotFound)
	if !notFoundErr {
		return err
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
			fileCtxs = []*config.Context{}
			logger.Debug("no device configuration config files found")
		}

	case policies.DeviceConfigFileProhibited:
		// If the device config file is prohibited, we will log a warning
		// if a file is found, but we will ultimately not fail. Instead, we
		// will just pass along an empty config.
		if err == nil && len(fileCtxs) > 0 {
			logger.Warn(
				"device config file(s) found, but its use is prohibited via policy. " +
					"the device config files will be ignored.",
			)
			fileCtxs = []*config.Context{}
		}

	default:
		return errors.NewPolicyViolationError(
			deviceFilePolicy.String(),
			"unsupported device config file policy",
		)
	}

	// Now, we can append whatever config contexts we got from file to the slice of all
	// device config contexts.
	deviceCtxs = append(deviceCtxs, fileCtxs...)

	var dynamicCtxs []*config.Context

	// Get device configs from dynamic registration
	dynamicCfgs, err := Context.dynamicDeviceConfigRegistrar(config.Plugin.DynamicRegistration.Config)

	// If the error is not a "config not found" error, then we will return it.
	_, notFoundErr = err.(*errors.ConfigsNotFound)
	if !notFoundErr {
		return err
	}

	for _, cfg := range dynamicCfgs {
		dynamicCtxs = append(dynamicCtxs, config.NewConfigContext("dynamic registration", cfg))
	}

	switch deviceDynamicPolicy {
	case policies.DeviceConfigDynamicRequired:
		if err != nil {
			return errors.NewPolicyViolationError(
				deviceDynamicPolicy.String(),
				fmt.Sprintf("dynamic device config(s) required, but none found: %v", err),
			)
		}

	case policies.DeviceConfigDynamicOptional:
		if err != nil {
			dynamicCtxs = []*config.Context{}
			logger.Debug("no dynamic device configuration(s) found")
		}

	case policies.DeviceConfigDynamicProhibited:
		// If dynamic device configs are prohibited, we will log a warning
		// if any are found, but we will ultimately not fail. Instead, we
		// will just pass along an empty config.
		if err == nil && len(dynamicCfgs) > 0 {
			logger.Warn(
				"dynamic device config(s) found, but its use is prohibited via policy. " +
					"the device config(s) will be ignored.",
			)
			dynamicCtxs = []*config.Context{}
		}

	default:
		return errors.NewPolicyViolationError(
			deviceDynamicPolicy.String(),
			"unsupported dynamic device config policy",
		)
	}

	// Now, we can append whatever config contexts we got from dynamic registration to the slice
	// of all device config contexts.
	deviceCtxs = append(deviceCtxs, dynamicCtxs...)

	// Validate the device configs
	for _, ctx := range deviceCtxs {
		// Validate config scheme
		multiErr := config.Validator.Validate(ctx)
		if multiErr.HasErrors() {
			return multiErr
		}
	}

	// Unify all the device configs
	unifiedCtx, err := config.UnifyDeviceConfigs(deviceCtxs)
	if err != nil {
		return err
	}

	// With the config validated and unified, we can now assign it to the global Device variable.
	config.Device = unifiedCtx.Config.(*config.DeviceConfig)
	return nil
}

// processPluginConfig searches for, reads, and validates the plugin configuration.
// Its behavior will vary depending on the plugin config policy that is set. If
// plugin config is processed successfully, it will be set to the global Plugin
// variable.
func processPluginConfig() error { // nolint: gocyclo
	// Get the plugin's policy for plugin config files.
	pluginFilePolicy := policies.GetPluginConfigFilePolicy()
	logger.Debugf("plugin config file policy: %s", pluginFilePolicy.String())

	// Now, try getting the plugin config from file.
	pluginCtx, err := config.GetPluginConfigFromFile()

	// If the error is not a "config not found" error, then we will return it.
	_, notFoundErr := err.(*errors.ConfigsNotFound)
	if !notFoundErr {
		return err
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
			ctx, e := config.NewDefaultPluginConfig()
			if e != nil {
				return e
			}
			pluginCtx = config.NewConfigContext("default", ctx)
		}

	case policies.PluginConfigFileProhibited:
		// If the plugin config file is prohibited, we will log a warning
		// if a file is found, but we will ultimately not fail. Instead, we
		// will just pass along an empty config.
		//
		// It is up to the user to specify the config (whether default of not)
		// when the plugin config is prohibited.
		if err == nil && pluginCtx != nil {
			logger.Warn(
				"plugin config file found, but its use is prohibited via policy. " +
					"you must ensure that the plugin has its config set manually.",
			)
			// The user should have specified the config, so we will take
			// that config and wrap it in a context for validation.
			pluginCtx = config.NewConfigContext("user defined", config.Plugin)
		}

	default:
		return errors.NewPolicyViolationError(
			pluginFilePolicy.String(),
			"unsupported plugin config file policy",
		)
	}

	// Validate the plugin config
	multiErr := config.Validator.Validate(pluginCtx)
	if multiErr.HasErrors() {
		return multiErr
	}

	// With the config validated, we can now assign it to the global Plugin variable.
	config.Plugin = pluginCtx.Config.(*config.PluginConfig)
	return nil
}

// processOutputTypeConfig searches for, reads, and validates the output type
// configuration from file. Its behavior will vary depending on the output type
// config policy that is set. If output type config is processed successfully,
// the found output type configs are returned.
func processOutputTypeConfig() ([]*config.OutputType, error) { // nolint: gocyclo
	// Get the plugin's policy for output type config files.
	outputTypeFilePolicy := policies.GetTypeConfigFilePolicy()
	logger.Debugf("output type config file policy: %s", outputTypeFilePolicy.String())

	// Now, try getting the output type config(s) from file.
	outputTypeCtxs, err := config.GetOutputTypeConfigsFromFile()

	// If the error is not a "config not found" error, then we will return it.
	_, notFoundErr := err.(*errors.ConfigsNotFound)
	if !notFoundErr {
		return nil, err
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
			outputTypeCtxs = []*config.Context{}
			logger.Debug("no type configuration config files found")
		}

	case policies.TypeConfigFileProhibited:
		// If the output type config file is prohibited, we will log a warning
		// if a file is found, but we will ultimately not fail. Instead, we
		// will just pass along an empty config.
		if err == nil && len(outputTypeCtxs) > 0 {
			logger.Warn(
				"output type config file(s) found, but its use is prohibited via policy. " +
					"the output type config files will be ignored.",
			)
			outputTypeCtxs = []*config.Context{}
		}

	default:
		return nil, errors.NewPolicyViolationError(
			outputTypeFilePolicy.String(),
			"unsupported output type config file policy",
		)
	}

	var outputs []*config.OutputType

	// Validate the plugin config
	for _, outputTypeCtx := range outputTypeCtxs {
		multiErr := config.Validator.Validate(outputTypeCtx)
		if multiErr.HasErrors() {
			return nil, multiErr
		}
		cfg := outputTypeCtx.Config.(*config.OutputType)
		outputs = append(outputs, cfg)
	}
	return outputs, nil
}
