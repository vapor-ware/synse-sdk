package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/internal/test"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/policies"
)

// TestNewConfigContext tests creating a new Context.
func TestNewConfigContext(t *testing.T) {
	var testTable = []struct {
		desc   string
		source string
		config ConfigBase
	}{
		{
			desc:   "Config is a pointer to a DeviceConfig",
			source: "test",
			config: &DeviceConfig{},
		},
		{
			desc:   "Config is a pointer to a PluginConfig",
			source: "test",
			config: &PluginConfig{},
		},
	}

	for _, testCase := range testTable {
		ctx := NewConfigContext(testCase.source, testCase.config)
		assert.NotNil(t, ctx, testCase.desc)
		assert.IsType(t, &ConfigContext{}, ctx, testCase.desc)
		assert.Equal(t, testCase.source, ctx.Source, testCase.desc)
		assert.Equal(t, testCase.config, ctx.Config, testCase.desc)
	}
}

// TestConfigContext_IsDeviceConfig tests whether the Config member is a DeviceConfig.
func TestConfigContext_IsDeviceConfig(t *testing.T) {
	var testTable = []struct {
		desc     string
		isDevCfg bool
		config   ConfigBase
	}{
		{
			desc:     "Config is a pointer to a DeviceConfig",
			isDevCfg: true,
			config:   &DeviceConfig{},
		},
		{
			desc:     "Config is a pointer to a PluginConfig",
			isDevCfg: false,
			config:   &PluginConfig{},
		},
		{
			desc:     "Config is a pointer to an OutputType",
			isDevCfg: false,
			config:   &OutputType{},
		},
	}

	for _, testCase := range testTable {
		ctx := ConfigContext{
			Source: "test",
			Config: testCase.config,
		}

		actual := ctx.IsDeviceConfig()
		assert.Equal(t, testCase.isDevCfg, actual, testCase.desc)
	}
}

// TestConfigContext_IsPluginConfig tests whether the Config member is a PluginConfig.
func TestConfigContext_IsPluginConfig(t *testing.T) {
	var testTable = []struct {
		desc        string
		isPluginCfg bool
		config      ConfigBase
	}{
		{
			desc:        "Config is a pointer to a DeviceConfig",
			isPluginCfg: false,
			config:      &DeviceConfig{},
		},
		{
			desc:        "Config is a pointer to a PluginConfig",
			isPluginCfg: true,
			config:      &PluginConfig{},
		},
		{
			desc:        "Config is a pointer to an OutputType",
			isPluginCfg: false,
			config:      &OutputType{},
		},
	}

	for _, testCase := range testTable {
		ctx := ConfigContext{
			Source: "test",
			Config: testCase.config,
		}

		actual := ctx.IsPluginConfig()
		assert.Equal(t, testCase.isPluginCfg, actual, testCase.desc)
	}
}

// TestConfigContext_IsOutputTypeConfig tests whether the Config member is an OutputType config.
func TestConfigContext_IsOutputTypeConfig(t *testing.T) {
	var testTable = []struct {
		desc         string
		isOutputType bool
		config       ConfigBase
	}{
		{
			desc:         "Config is a pointer to an OutputType",
			isOutputType: true,
			config:       &OutputType{},
		},
		{
			desc:         "Config is a pointer to a PluginConfig",
			isOutputType: false,
			config:       &PluginConfig{},
		},
		{
			desc:         "Config is a pointer to a DeviceConfig",
			isOutputType: false,
			config:       &DeviceConfig{},
		},
	}

	for _, testCase := range testTable {
		ctx := ConfigContext{
			Source: "test",
			Config: testCase.config,
		}

		actual := ctx.IsOutputTypeConfig()
		assert.Equal(t, testCase.isOutputType, actual, testCase.desc)
	}
}

// TestNewSchemeVersion_Ok tests creating a new Version with no errors.
func TestNewSchemeVersion_Ok(t *testing.T) {
	var testTable = []struct {
		desc     string
		in       string
		expected ConfigVersion
	}{
		{
			desc:     "Version with only major specified",
			in:       "1",
			expected: ConfigVersion{1, 0},
		},
		{
			desc:     "Version with major and 0-valued minor",
			in:       "1.0",
			expected: ConfigVersion{1, 0},
		},
		{
			desc:     "Version with 0-valued major and minor",
			in:       "0.1",
			expected: ConfigVersion{0, 1},
		},
		{
			desc:     "Version with non-0 major and minor",
			in:       "2.5",
			expected: ConfigVersion{2, 5},
		},
		{
			desc:     "Version with large major/minor",
			in:       "12345.12345",
			expected: ConfigVersion{12345, 12345},
		},
		{
			desc:     "Version with double zero major",
			in:       "00.1",
			expected: ConfigVersion{0, 1},
		},
		{
			desc:     "Version with double zero minor",
			in:       "1.00",
			expected: ConfigVersion{1, 0},
		},
	}

	for _, testCase := range testTable {
		sv, err := NewVersion(testCase.in)
		assert.NoError(t, err, testCase.desc)
		assert.NotNil(t, sv, testCase.desc)
		assert.IsType(t, &ConfigVersion{}, sv, testCase.desc)
		assert.Equal(t, testCase.expected.Major, sv.Major, testCase.desc)
		assert.Equal(t, testCase.expected.Minor, sv.Minor, testCase.desc)
	}
}

// TestNewSchemeVersion_Error tests creating a new Version with errors.
func TestNewSchemeVersion_Error(t *testing.T) {
	var testTable = []struct {
		desc string
		in   string
	}{
		{
			desc: "Empty string used as version",
			in:   "",
		},
		{
			desc: "Invalid major version, no minor (not an int)",
			in:   "xyz",
		},
		{
			desc: "Invalid major version (not an int)",
			in:   "xyz.0",
		},
		{
			desc: "Invalid minor version (not an int)",
			in:   "1.xyz",
		},
		{
			desc: "Invalid major and minor versions (not an int)",
			in:   "xyz.xyz",
		},
		{
			desc: "Extra version number components",
			in:   "1.2.3.4",
		},
	}

	for _, testCase := range testTable {
		sv, err := NewVersion(testCase.in)
		assert.Nil(t, sv, testCase.desc)
		assert.Error(t, err, testCase.desc)
	}
}

// TestSchemeVersion_String tests converting a Version to a string
func TestSchemeVersion_String(t *testing.T) {
	var testTable = []struct {
		scheme   ConfigVersion
		expected string
	}{
		{
			scheme:   ConfigVersion{0, 1},
			expected: "0.1",
		},
		{
			scheme:   ConfigVersion{1, 0},
			expected: "1.0",
		},
		{
			scheme:   ConfigVersion{1, 1},
			expected: "1.1",
		},
		{
			scheme:   ConfigVersion{1234, 4321},
			expected: "1234.4321",
		},
	}

	for _, testCase := range testTable {
		actual := testCase.scheme.String()
		assert.Equal(t, testCase.expected, actual)
	}
}

// TestSchemeVersion_IsEqual test equality of SchemeVersions
func TestSchemeVersion_IsEqual(t *testing.T) {
	var testTable = []struct {
		scheme1 *ConfigVersion
		scheme2 *ConfigVersion
		equal   bool
	}{
		{
			scheme1: &ConfigVersion{1, 0},
			scheme2: &ConfigVersion{1, 0},
			equal:   true,
		},
		{
			scheme1: &ConfigVersion{0, 1},
			scheme2: &ConfigVersion{0, 1},
			equal:   true,
		},
		{
			scheme1: &ConfigVersion{4, 51},
			scheme2: &ConfigVersion{4, 51},
			equal:   true,
		},
		{
			scheme1: &ConfigVersion{1, 0},
			scheme2: &ConfigVersion{2, 0},
			equal:   false,
		},
		{
			scheme1: &ConfigVersion{1, 1},
			scheme2: &ConfigVersion{1, 2},
			equal:   false,
		},
	}

	for _, testCase := range testTable {
		actual := testCase.scheme1.IsEqual(testCase.scheme2)
		assert.Equal(t, testCase.equal, actual)
	}
}

// TestSchemeVersion_IsLessThan tests if one Version is less than another
func TestSchemeVersion_IsLessThan(t *testing.T) {
	var testTable = []struct {
		scheme1  *ConfigVersion
		scheme2  *ConfigVersion
		lessThan bool
	}{
		{
			scheme1:  &ConfigVersion{1, 0},
			scheme2:  &ConfigVersion{1, 0},
			lessThan: false,
		},
		{
			scheme1:  &ConfigVersion{0, 1},
			scheme2:  &ConfigVersion{0, 1},
			lessThan: false,
		},
		{
			scheme1:  &ConfigVersion{4, 51},
			scheme2:  &ConfigVersion{4, 51},
			lessThan: false,
		},
		{
			scheme1:  &ConfigVersion{1, 0},
			scheme2:  &ConfigVersion{2, 0},
			lessThan: true,
		},
		{
			scheme1:  &ConfigVersion{1, 1},
			scheme2:  &ConfigVersion{1, 2},
			lessThan: true,
		},
		{
			scheme1:  &ConfigVersion{1, 2},
			scheme2:  &ConfigVersion{1, 1},
			lessThan: false,
		},
	}

	for _, testCase := range testTable {
		actual := testCase.scheme1.IsLessThan(testCase.scheme2)
		assert.Equal(t, testCase.lessThan, actual)
	}
}

// TestSchemeVersion_IsGreaterOrEqualTo tests if one Version is greater than
// or qual to another
func TestSchemeVersion_IsGreaterOrEqualTo(t *testing.T) {
	var testTable = []struct {
		scheme1 *ConfigVersion
		scheme2 *ConfigVersion
		gte     bool
	}{
		{
			scheme1: &ConfigVersion{1, 0},
			scheme2: &ConfigVersion{1, 0},
			gte:     true,
		},
		{
			scheme1: &ConfigVersion{0, 1},
			scheme2: &ConfigVersion{0, 1},
			gte:     true,
		},
		{
			scheme1: &ConfigVersion{4, 51},
			scheme2: &ConfigVersion{4, 51},
			gte:     true,
		},
		{
			scheme1: &ConfigVersion{1, 0},
			scheme2: &ConfigVersion{2, 0},
			gte:     false,
		},
		{
			scheme1: &ConfigVersion{1, 1},
			scheme2: &ConfigVersion{1, 2},
			gte:     false,
		},
		{
			scheme1: &ConfigVersion{1, 2},
			scheme2: &ConfigVersion{1, 1},
			gte:     true,
		},
		{
			scheme1: &ConfigVersion{2, 1},
			scheme2: &ConfigVersion{1, 2},
			gte:     true,
		},
	}

	for _, testCase := range testTable {
		actual := testCase.scheme1.IsGreaterOrEqualTo(testCase.scheme2)
		assert.Equal(t, testCase.gte, actual)
	}
}

// TestConfigVersion_GetSchemeVersion_Ok tests getting the scheme version from a SchemeVersion
func TestConfigVersion_GetSchemeVersion_Ok(t *testing.T) {
	var testTable = []struct {
		desc    string
		version string
		scheme  ConfigVersion
	}{
		{
			desc:    "Version with only major specified",
			version: "1",
			scheme:  ConfigVersion{1, 0},
		},
		{
			desc:    "Version with major and 0-valued minor",
			version: "1.0",
			scheme:  ConfigVersion{1, 0},
		},
		{
			desc:    "Version with 0-valued major and minor",
			version: "0.1",
			scheme:  ConfigVersion{0, 1},
		},
		{
			desc:    "Version with non-0 major and minor",
			version: "2.5",
			scheme:  ConfigVersion{2, 5},
		},
		{
			desc:    "Version with large major/minor",
			version: "12345.12345",
			scheme:  ConfigVersion{12345, 12345},
		},
		{
			desc:    "Version with double zero major",
			version: "00.1",
			scheme:  ConfigVersion{0, 1},
		},
		{
			desc:    "Version with double zero minor",
			version: "1.00",
			scheme:  ConfigVersion{1, 0},
		},
	}

	for _, testCase := range testTable {
		cfgVer := SchemeVersion{Version: testCase.version}
		sv, err := cfgVer.GetVersion()
		assert.NoError(t, err, testCase.desc)
		assert.Equal(t, testCase.scheme.Major, sv.Major, testCase.desc)
		assert.Equal(t, testCase.scheme.Minor, sv.Minor, testCase.desc)
	}
}

// TestConfigVersion_GetSchemeVersion_Error tests getting the scheme version from a SchemeVersion
// which results in error
func TestConfigVersion_GetSchemeVersion_Error(t *testing.T) {
	var testTable = []struct {
		desc    string
		version string
	}{
		{
			desc:    "Empty string used as version",
			version: "",
		},
		{
			desc:    "Invalid major version, no minor (not an int)",
			version: "xyz",
		},
		{
			desc:    "Invalid major version (not an int)",
			version: "xyz.0",
		},
		{
			desc:    "Invalid minor version (not an int)",
			version: "1.xyz",
		},
		{
			desc:    "Invalid major and minor versions (not an int)",
			version: "xyz.xyz",
		},
		{
			desc:    "Extra version number components",
			version: "1.2.3.4",
		},
	}

	for _, testCase := range testTable {
		cfgVer := SchemeVersion{Version: testCase.version}
		sv, err := cfgVer.GetVersion()
		assert.Error(t, err, testCase.desc)
		assert.Nil(t, sv, testCase.desc)
	}
}

// TestUnifyDeviceConfigs_NoConfigs tests unifying configs when no
// configs are given.
func TestUnifyDeviceConfigs_NoConfigs(t *testing.T) {
	ctx, err := unifyDeviceConfigs([]*ConfigContext{})
	assert.Error(t, err)
	assert.Nil(t, ctx)
}

// TestUnifyDeviceConfigs_NoDeviceConfig tests unifying configs when the
// contexts specified do not contain DeviceConfigs.
func TestUnifyDeviceConfigs_NoDeviceConfig(t *testing.T) {
	ctx, err := unifyDeviceConfigs([]*ConfigContext{
		{
			Source: "test",
			Config: &PluginConfig{},
		},
	})

	assert.Error(t, err)
	assert.Nil(t, ctx)
}

// TestUnifyDeviceConfigs tests unifying configs when there is only one config
// to unify.
func TestUnifyDeviceConfigs(t *testing.T) {
	ctx, err := unifyDeviceConfigs([]*ConfigContext{
		{
			Source: "test",
			Config: &DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations: []*LocationConfig{
					{
						Name:  "test",
						Rack:  &LocationData{Name: "rack"},
						Board: &LocationData{Name: "board"},
					},
				},
				Devices: []*DeviceKind{
					{Name: "test-device"},
				},
			},
		},
	})

	assert.NoError(t, err)
	assert.True(t, ctx.IsDeviceConfig())
	cfg := ctx.Config.(*DeviceConfig)
	assert.Equal(t, 1, len(cfg.Devices))
	assert.Equal(t, 1, len(cfg.Locations))
}

// TestUnifyDeviceConfigs2 tests unifying configs when there are multiple
// configs to unify.
func TestUnifyDeviceConfigs2(t *testing.T) {
	ctx, err := unifyDeviceConfigs([]*ConfigContext{
		{
			Source: "test",
			Config: &DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations: []*LocationConfig{
					{
						Name:  "loc-1",
						Rack:  &LocationData{Name: "rack"},
						Board: &LocationData{Name: "board"},
					},
				},
				Devices: []*DeviceKind{
					{Name: "test-device-1"},
				},
			},
		},
		{
			Source: "test",
			Config: &DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations: []*LocationConfig{
					{
						Name:  "loc-2",
						Rack:  &LocationData{Name: "rack"},
						Board: &LocationData{Name: "board"},
					},
					{
						Name:  "loc-3",
						Rack:  &LocationData{Name: "rack"},
						Board: &LocationData{Name: "board"},
					},
				},
				Devices: []*DeviceKind{
					{Name: "test-device-2"},
				},
			},
		},
		{
			Source: "test",
			Config: &DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations: []*LocationConfig{
					{
						Name:  "loc-4",
						Rack:  &LocationData{Name: "rack"},
						Board: &LocationData{Name: "board"},
					},
				},
				Devices: []*DeviceKind{
					{Name: "test-device-3"},
					{Name: "test-device-4"},
				},
			},
		},
	})

	assert.NoError(t, err)
	assert.True(t, ctx.IsDeviceConfig())
	cfg := ctx.Config.(*DeviceConfig)
	assert.Equal(t, 4, len(cfg.Devices))
	assert.Equal(t, 4, len(cfg.Locations))
}

// Test_processOutputTypeConfig_None_Optional tests getting output type config from file when
// no files are found and the policy is optional.
func Test_processOutputTypeConfig_None_Optional(t *testing.T) {
	defer func() {
		resetContext()
		policies.Clear()
	}()

	policies.Add(policies.TypeConfigFileOptional)

	outputs, err := processOutputTypeConfig()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(outputs))
}

// Test_processOutputTypeConfig_None_Required tests getting output type config from file when
// no files are found and the policy is required.
func Test_processOutputTypeConfig_None_Required(t *testing.T) {
	defer func() {
		resetContext()
		policies.Clear()
	}()

	policies.Add(policies.TypeConfigFileRequired)

	outputs, err := processOutputTypeConfig()
	assert.Error(t, err)
	assert.IsType(t, &errors.PolicyViolationError{}, err)
	assert.Nil(t, outputs)
}

// Test_processOutputTypeConfig_None_Prohibited tests getting output type config from file when
// no files are found and the policy is prohibited.
func Test_processOutputTypeConfig_None_Prohibited(t *testing.T) {
	defer func() {
		resetContext()
		policies.Clear()
	}()

	policies.Add(policies.TypeConfigFileProhibited)

	outputs, err := processOutputTypeConfig()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(outputs))
}

// Test_processOutputTypeConfig_One_Optional tests getting output type config from file when
// one file is found and the policy is optional.
func Test_processOutputTypeConfig_One_Optional(t *testing.T) {
	test.SetEnv(t, EnvOutputTypeConfig, "testdata/output_type/ok.yml")
	defer func() {
		test.RemoveEnv(t, EnvOutputTypeConfig)
		resetContext()
		policies.Clear()
	}()

	policies.Add(policies.TypeConfigFileOptional)

	outputs, err := processOutputTypeConfig()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(outputs))
}

// Test_processOutputTypeConfig_One_Required tests getting output type config from file when
// one file is found and the policy is required.
func Test_processOutputTypeConfig_One_Required(t *testing.T) {
	test.SetEnv(t, EnvOutputTypeConfig, "testdata/output_type/ok.yml")
	defer func() {
		test.RemoveEnv(t, EnvOutputTypeConfig)
		resetContext()
		policies.Clear()
	}()

	policies.Add(policies.TypeConfigFileRequired)

	outputs, err := processOutputTypeConfig()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(outputs))
}

// Test_processOutputTypeConfig_One_Prohibited tests getting output type config from file when
// one file is found and the policy is prohibited.
func Test_processOutputTypeConfig_One_Prohibited(t *testing.T) {
	test.SetEnv(t, EnvOutputTypeConfig, "testdata/output_type/ok.yml")
	defer func() {
		test.RemoveEnv(t, EnvOutputTypeConfig)
		resetContext()
		policies.Clear()
	}()

	policies.Add(policies.TypeConfigFileProhibited)

	outputs, err := processOutputTypeConfig()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(outputs))
}

// Test_processOutputTypeConfig_withErrors tests getting output type configs when the
// configs have validation errors.
func Test_processOutputTypeConfig_withErrors(t *testing.T) {
	test.SetEnv(t, EnvOutputTypeConfig, "testdata/output_type/invalid.yml")
	defer func() {
		test.RemoveEnv(t, EnvOutputTypeConfig)
		resetContext()
		policies.Clear()
	}()

	policies.Add(policies.TypeConfigFileOptional)

	outputs, err := processOutputTypeConfig()
	assert.Error(t, err)
	assert.Nil(t, outputs)
}

// Test_processOutputTypeConfig_withErrors2 tests getting output type configs when there
// is an error finding configs.
func Test_processOutputTypeConfig_withErrors2(t *testing.T) {
	test.SetEnv(t, EnvOutputTypeConfig, "testdata/output_type/nonexistent.yml")
	defer func() {
		test.RemoveEnv(t, EnvOutputTypeConfig)
		resetContext()
		policies.Clear()
	}()

	policies.Add(policies.TypeConfigFileOptional)

	outputs, err := processOutputTypeConfig()
	assert.Error(t, err)
	assert.Nil(t, outputs)
}

// Test_processPluginConfig_None_Optional tests getting plugin config from file when
// no files are found and the policy is optional.
func Test_processPluginConfig_None_Optional(t *testing.T) {
	defer func() {
		resetContext()
		policies.Clear()
		Config.reset()
	}()

	policies.Add(policies.PluginConfigFileOptional)

	assert.Nil(t, Config.Plugin)
	err := processPluginConfig()
	assert.NoError(t, err)
	assert.NotNil(t, Config.Plugin) // default gets set
}

// Test_processPluginConfig_None_Required tests getting plugin config from file when
// no files are found and the policy is required.
func Test_processPluginConfig_None_Required(t *testing.T) {
	defer func() {
		resetContext()
		policies.Clear()
		Config.reset()
	}()

	policies.Add(policies.PluginConfigFileRequired)

	assert.Nil(t, Config.Plugin)
	err := processPluginConfig()
	assert.Error(t, err)
	assert.IsType(t, &errors.PolicyViolationError{}, err)
	assert.Nil(t, Config.Plugin)
}

// Test_processPluginConfig_None_Prohibited tests getting plugin config from file when
// no files are found and the policy is prohibited.
func Test_processPluginConfig_None_Prohibited(t *testing.T) {
	defer func() {
		resetContext()
		policies.Clear()
		Config.reset()
	}()

	policies.Add(policies.PluginConfigFileProhibited)

	assert.Nil(t, Config.Plugin)
	err := processPluginConfig()
	assert.Error(t, err)
	assert.IsType(t, &errors.PolicyViolationError{}, err)
	assert.Nil(t, Config.Plugin)
}

// Test_processPluginConfig_None_Prohibited2 tests getting plugin config from file when
// no files are found and the policy is prohibited.
func Test_processPluginConfig_None_Prohibited2(t *testing.T) {
	defer func() {
		resetContext()
		policies.Clear()
		Config.reset()
	}()

	policies.Add(policies.PluginConfigFileProhibited)

	// here we set the plugin config manually since it is prohibited via file
	Config.Plugin = &PluginConfig{
		SchemeVersion: SchemeVersion{Version: "1.0"},
		Network: &NetworkSettings{
			Type:    "tcp",
			Address: "foo.bar.baz",
		},
	}

	assert.NotNil(t, Config.Plugin)
	err := processPluginConfig()
	assert.NoError(t, err)
	assert.NotNil(t, Config.Plugin)
}

// Test_processPluginConfig_One_Optional tests getting plugin config from file when
// one file is found and the policy is optional.
func Test_processPluginConfig_One_Optional(t *testing.T) {
	test.SetEnv(t, EnvPluginConfig, "testdata/plugin/ok/config.yml")
	defer func() {
		test.RemoveEnv(t, EnvPluginConfig)
		resetContext()
		policies.Clear()
		Config.reset()
	}()

	policies.Add(policies.PluginConfigFileOptional)

	assert.Nil(t, Config.Plugin)
	err := processPluginConfig()
	assert.NoError(t, err)
	assert.NotNil(t, Config.Plugin)
}

// Test_processPluginConfig_One_Required tests getting plugin config from file when
// one file is found and the policy is required.
func Test_processPluginConfig_One_Required(t *testing.T) {
	test.SetEnv(t, EnvPluginConfig, "testdata/plugin/ok/config.yml")
	defer func() {
		test.RemoveEnv(t, EnvPluginConfig)
		resetContext()
		policies.Clear()
		Config.reset()
	}()

	policies.Add(policies.PluginConfigFileRequired)

	assert.Nil(t, Config.Plugin)
	err := processPluginConfig()
	assert.NoError(t, err)
	assert.NotNil(t, Config.Plugin)
}

// Test_processPluginConfig_One_Prohibited tests getting plugin config from file when
// one file is found and the policy is prohibited.
func Test_processPluginConfig_One_Prohibited(t *testing.T) {
	test.SetEnv(t, EnvPluginConfig, "testdata/plugin/ok/config.yml")
	defer func() {
		test.RemoveEnv(t, EnvPluginConfig)
		resetContext()
		policies.Clear()
		Config.reset()
	}()

	policies.Add(policies.PluginConfigFileProhibited)

	// here we set the plugin config manually since it is prohibited via file
	Config.Plugin = &PluginConfig{
		SchemeVersion: SchemeVersion{Version: "1.0"},
		Network: &NetworkSettings{
			Type:    "tcp",
			Address: "foo.bar.baz",
		},
	}

	assert.NotNil(t, Config.Plugin)
	err := processPluginConfig()
	assert.NoError(t, err)
	assert.NotNil(t, Config.Plugin)
}

// Test_processPluginConfig_withErrors tests getting plugin config when the
// configs have validation errors.
func Test_processPluginConfig_withErrors(t *testing.T) {
	test.SetEnv(t, EnvPluginConfig, "testdata/plugin/invalid/config.yml")
	defer func() {
		test.RemoveEnv(t, EnvPluginConfig)
		resetContext()
		policies.Clear()
	}()

	policies.Add(policies.PluginConfigFileOptional)

	assert.Nil(t, Config.Plugin)
	err := processPluginConfig()
	assert.Error(t, err)
	assert.Nil(t, Config.Plugin)
}

// Test_processPluginConfig_withErrors2 tests getting plugin config when there
// is an error finding configs.
func Test_processPluginConfig_withErrors2(t *testing.T) {
	test.SetEnv(t, EnvPluginConfig, "testdata/output_type/nonexistent/config.yml")
	defer func() {
		test.RemoveEnv(t, EnvPluginConfig)
		resetContext()
		policies.Clear()
	}()

	policies.Add(policies.PluginConfigFileOptional)

	assert.Nil(t, Config.Plugin)
	err := processPluginConfig()
	assert.Error(t, err)
	assert.Nil(t, Config.Plugin)
}

// Test_processDeviceConfigs_File_None_Optional tests getting device config(s) from file when
// no files are found and the policy is optional.
func Test_processDeviceConfigs_File_None_Optional(t *testing.T) {
	defer func() {
		resetContext()
		policies.Clear()
		Config.reset()
	}()

	Config.Plugin = &PluginConfig{
		DynamicRegistration: &DynamicRegistrationSettings{
			Config: map[string]interface{}{},
		},
	}

	policies.Add(policies.DeviceConfigFileOptional)
	policies.Add(policies.DeviceConfigDynamicOptional)

	assert.Nil(t, Config.Device)
	err := processDeviceConfigs()
	assert.NoError(t, err)
	assert.NotNil(t, Config.Device)
	assert.Equal(t, 0, len(Config.Device.Devices))
}

// Test_processDeviceConfigs_File_None_Required tests getting device config(s) from file when
// no files are found and the policy is required.
func Test_processDeviceConfigs_File_None_Required(t *testing.T) {
	defer func() {
		resetContext()
		policies.Clear()
		Config.reset()
	}()

	Config.Plugin = &PluginConfig{
		DynamicRegistration: &DynamicRegistrationSettings{
			Config: map[string]interface{}{},
		},
	}

	policies.Add(policies.DeviceConfigFileRequired)
	policies.Add(policies.DeviceConfigDynamicOptional)

	assert.Nil(t, Config.Device)
	err := processDeviceConfigs()
	assert.Error(t, err)
	assert.IsType(t, &errors.PolicyViolationError{}, err)
	assert.Nil(t, Config.Device)
}

// Test_processDeviceConfigs_File_None_Prohibited tests getting device config(s) from file when
// no files are found and the policy is prohibited.
func Test_processDeviceConfigs_File_None_Prohibited(t *testing.T) {
	defer func() {
		resetContext()
		policies.Clear()
		Config.reset()
	}()

	Config.Plugin = &PluginConfig{
		DynamicRegistration: &DynamicRegistrationSettings{
			Config: map[string]interface{}{},
		},
	}

	policies.Add(policies.DeviceConfigFileProhibited)
	policies.Add(policies.DeviceConfigDynamicOptional)

	assert.Nil(t, Config.Device)
	err := processDeviceConfigs()
	assert.NoError(t, err)
	assert.NotNil(t, Config.Device)
	assert.Equal(t, 0, len(Config.Device.Devices))
}

// Test_processDeviceConfigs_Dynamic_None_Optional tests getting device config(s) from dynamic
// registration when nothing is returned and the policy is optional.
func Test_processDeviceConfigs_Dynamic_None_Optional(t *testing.T) {
	defer func() {
		resetContext()
		policies.Clear()
		Config.reset()
	}()

	Config.Plugin = &PluginConfig{
		DynamicRegistration: &DynamicRegistrationSettings{
			Config: map[string]interface{}{},
		},
	}

	ctx.dynamicDeviceConfigRegistrar = func(i map[string]interface{}) ([]*DeviceConfig, error) {
		return nil, errors.NewConfigsNotFoundError([]string{"test"})
	}

	policies.Add(policies.DeviceConfigFileOptional)
	policies.Add(policies.DeviceConfigDynamicOptional)

	assert.Nil(t, Config.Device)
	err := processDeviceConfigs()
	assert.NoError(t, err)
	assert.NotNil(t, Config.Device)
	assert.Equal(t, 0, len(Config.Device.Devices))
}

// Test_processDeviceConfigs_Dynamic_None_Required tests getting device config(s) from dynamic
// registration when nothing is returned and the policy is required.
func Test_processDeviceConfigs_Dynamic_None_Required(t *testing.T) {
	defer func() {
		resetContext()
		policies.Clear()
		Config.reset()
	}()

	Config.Plugin = &PluginConfig{
		DynamicRegistration: &DynamicRegistrationSettings{
			Config: map[string]interface{}{},
		},
	}

	ctx.dynamicDeviceConfigRegistrar = func(i map[string]interface{}) ([]*DeviceConfig, error) {
		return nil, errors.NewConfigsNotFoundError([]string{"test"})
	}

	policies.Add(policies.DeviceConfigFileOptional)
	policies.Add(policies.DeviceConfigDynamicRequired)

	assert.Nil(t, Config.Device)
	err := processDeviceConfigs()
	assert.Error(t, err)
	assert.IsType(t, &errors.PolicyViolationError{}, err)
	assert.Nil(t, Config.Device)
}

// Test_processDeviceConfigs_Dynamic_None_Prohibited tests getting device config(s) from dynamic
// registration when nothing is returned and the policy is prohibited.
func Test_processDeviceConfigs_Dynamic_None_Prohibited(t *testing.T) {
	defer func() {
		resetContext()
		policies.Clear()
		Config.reset()
	}()

	Config.Plugin = &PluginConfig{
		DynamicRegistration: &DynamicRegistrationSettings{
			Config: map[string]interface{}{},
		},
	}

	ctx.dynamicDeviceConfigRegistrar = func(i map[string]interface{}) ([]*DeviceConfig, error) {
		return nil, errors.NewConfigsNotFoundError([]string{"test"})
	}

	policies.Add(policies.DeviceConfigFileOptional)
	policies.Add(policies.DeviceConfigDynamicProhibited)

	assert.Nil(t, Config.Device)
	err := processDeviceConfigs()
	assert.NoError(t, err)
	assert.NotNil(t, Config.Device)
	assert.Equal(t, 0, len(Config.Device.Devices))
}

// Test_processDeviceConfigs_File_One_Optional tests getting device config(s) from file when
// one file is found and the policy is optional.
func Test_processDeviceConfigs_File_One_Optional(t *testing.T) {
	test.SetEnv(t, EnvDeviceConfig, "testdata/device/ok.yml")
	defer func() {
		test.RemoveEnv(t, EnvDeviceConfig)
		resetContext()
		policies.Clear()
		Config.reset()
	}()

	Config.Plugin = &PluginConfig{
		DynamicRegistration: &DynamicRegistrationSettings{
			Config: map[string]interface{}{},
		},
	}

	policies.Add(policies.DeviceConfigFileOptional)
	policies.Add(policies.DeviceConfigDynamicOptional)

	assert.Nil(t, Config.Device)
	err := processDeviceConfigs()
	assert.NoError(t, err)
	assert.NotNil(t, Config.Device)
	assert.Equal(t, 1, len(Config.Device.Devices))
}

// Test_processDeviceConfigs_File_One_Required tests getting device config(s) from file when
// one file is found and the policy is required.
func Test_processDeviceConfigs_File_One_Required(t *testing.T) {
	test.SetEnv(t, EnvDeviceConfig, "testdata/device/ok.yml")
	defer func() {
		test.RemoveEnv(t, EnvDeviceConfig)
		resetContext()
		policies.Clear()
		Config.reset()
	}()

	Config.Plugin = &PluginConfig{
		DynamicRegistration: &DynamicRegistrationSettings{
			Config: map[string]interface{}{},
		},
	}

	policies.Add(policies.DeviceConfigFileRequired)
	policies.Add(policies.DeviceConfigDynamicOptional)

	assert.Nil(t, Config.Device)
	err := processDeviceConfigs()
	assert.NoError(t, err)
	assert.NotNil(t, Config.Device)
	assert.Equal(t, 1, len(Config.Device.Devices))
}

// Test_processDeviceConfigs_File_One_Prohibited tests getting device config(s) from file when
// one file is found and the policy is prohibited.
func Test_processDeviceConfigs_File_One_Prohibited(t *testing.T) {
	test.SetEnv(t, EnvDeviceConfig, "testdata/device/ok.yml")
	defer func() {
		test.RemoveEnv(t, EnvDeviceConfig)
		resetContext()
		policies.Clear()
		Config.reset()
	}()

	Config.Plugin = &PluginConfig{
		DynamicRegistration: &DynamicRegistrationSettings{
			Config: map[string]interface{}{},
		},
	}

	policies.Add(policies.DeviceConfigFileProhibited)
	policies.Add(policies.DeviceConfigDynamicOptional)

	assert.Nil(t, Config.Device)
	err := processDeviceConfigs()
	assert.NoError(t, err)
	assert.NotNil(t, Config.Device)
	assert.Equal(t, 0, len(Config.Device.Devices))
}

// Test_processDeviceConfigs_Dynamic_One_Optional tests getting device config(s) from dynamic
// registration when one config is returned and the policy is optional.
func Test_processDeviceConfigs_Dynamic_One_Optional(t *testing.T) {
	defer func() {
		resetContext()
		policies.Clear()
		Config.reset()
	}()

	Config.Plugin = &PluginConfig{
		DynamicRegistration: &DynamicRegistrationSettings{
			Config: map[string]interface{}{},
		},
	}

	ctx.dynamicDeviceConfigRegistrar = func(i map[string]interface{}) ([]*DeviceConfig, error) {
		return []*DeviceConfig{
			{
				SchemeVersion: SchemeVersion{Version: currentDeviceSchemeVersion},
				Locations: []*LocationConfig{{
					Name:  "foo",
					Rack:  &LocationData{Name: "rack"},
					Board: &LocationData{Name: "board"},
				}},
				Devices: []*DeviceKind{{
					Name: "test",
					Instances: []*DeviceInstance{{
						Location: "foo",
					}},
				}},
			},
		}, nil
	}

	policies.Add(policies.DeviceConfigFileOptional)
	policies.Add(policies.DeviceConfigDynamicOptional)

	assert.Nil(t, Config.Device)
	err := processDeviceConfigs()
	assert.NoError(t, err)
	assert.NotNil(t, Config.Device)
	assert.Equal(t, 1, len(Config.Device.Devices))
}

// Test_processDeviceConfigs_Dynamic_One_Required tests getting device config(s) from dynamic
// registration when one config is returned and the policy is required.
func Test_processDeviceConfigs_Dynamic_One_Required(t *testing.T) {
	defer func() {
		resetContext()
		policies.Clear()
		Config.reset()
	}()

	Config.Plugin = &PluginConfig{
		DynamicRegistration: &DynamicRegistrationSettings{
			Config: map[string]interface{}{},
		},
	}

	ctx.dynamicDeviceConfigRegistrar = func(i map[string]interface{}) ([]*DeviceConfig, error) {
		return []*DeviceConfig{
			{
				SchemeVersion: SchemeVersion{Version: currentDeviceSchemeVersion},
				Locations: []*LocationConfig{{
					Name:  "foo",
					Rack:  &LocationData{Name: "rack"},
					Board: &LocationData{Name: "board"},
				}},
				Devices: []*DeviceKind{{
					Name: "test",
					Instances: []*DeviceInstance{{
						Location: "foo",
					}},
				}},
			},
		}, nil
	}

	policies.Add(policies.DeviceConfigFileOptional)
	policies.Add(policies.DeviceConfigDynamicRequired)

	assert.Nil(t, Config.Device)
	err := processDeviceConfigs()
	assert.NoError(t, err)
	assert.NotNil(t, Config.Device)
	assert.Equal(t, 1, len(Config.Device.Devices))
}

// Test_processDeviceConfigs_Dynamic_One_Prohibited tests getting device config(s) from dynamic
// registration when one config is returned and the policy is prohibited.
func Test_processDeviceConfigs_Dynamic_One_Prohibited(t *testing.T) {
	defer func() {
		resetContext()
		policies.Clear()
		Config.reset()
	}()

	Config.Plugin = &PluginConfig{
		DynamicRegistration: &DynamicRegistrationSettings{
			Config: map[string]interface{}{},
		},
	}

	ctx.dynamicDeviceConfigRegistrar = func(i map[string]interface{}) ([]*DeviceConfig, error) {
		return []*DeviceConfig{
			{
				SchemeVersion: SchemeVersion{Version: currentDeviceSchemeVersion},
				Locations: []*LocationConfig{{
					Name:  "foo",
					Rack:  &LocationData{Name: "rack"},
					Board: &LocationData{Name: "board"},
				}},
				Devices: []*DeviceKind{{
					Name: "test",
					Instances: []*DeviceInstance{{
						Location: "foo",
					}},
				}},
			},
		}, nil
	}

	policies.Add(policies.DeviceConfigFileOptional)
	policies.Add(policies.DeviceConfigDynamicProhibited)

	assert.Nil(t, Config.Device)
	err := processDeviceConfigs()
	assert.NoError(t, err)
	assert.NotNil(t, Config.Device)
	assert.Equal(t, 0, len(Config.Device.Devices))
}

// Test_processDeviceConfig_withErrors tests getting device config when the
// configs have validation errors.
func Test_processDeviceConfig_withErrors(t *testing.T) {
	test.SetEnv(t, EnvDeviceConfig, "testdata/device/invalid.yml")
	defer func() {
		test.RemoveEnv(t, EnvDeviceConfig)
		resetContext()
		policies.Clear()
		Config.reset()
	}()

	Config.Plugin = &PluginConfig{
		DynamicRegistration: &DynamicRegistrationSettings{
			Config: map[string]interface{}{},
		},
	}

	policies.Add(policies.DeviceConfigFileOptional)
	policies.Add(policies.DeviceConfigDynamicOptional)

	assert.Nil(t, Config.Device)
	err := processDeviceConfigs()
	assert.Error(t, err)
	assert.Nil(t, Config.Device)
}

// Test_processDeviceConfig_withErrors2 tests getting device config when there
// is an error finding configs.
func Test_processDeviceConfig_withErrors2(t *testing.T) {
	test.SetEnv(t, EnvDeviceConfig, "testdata/device/nonexistent.yml")
	defer func() {
		test.RemoveEnv(t, EnvDeviceConfig)
		resetContext()
		policies.Clear()
		Config.reset()
	}()

	Config.Plugin = &PluginConfig{
		DynamicRegistration: &DynamicRegistrationSettings{
			Config: map[string]interface{}{},
		},
	}

	policies.Add(policies.DeviceConfigFileOptional)
	policies.Add(policies.DeviceConfigDynamicOptional)

	assert.Nil(t, Config.Device)
	err := processDeviceConfigs()
	assert.Error(t, err)
	assert.Nil(t, Config.Device)
}
