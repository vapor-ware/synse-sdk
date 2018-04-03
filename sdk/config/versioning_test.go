package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConfigVersion_ToString tests converting ConfigVersion to a string.
func TestConfigVersion_ToString(t *testing.T) {
	var cases = []struct {
		in  configVersion
		out string
	}{
		{
			in:  configVersion{0, 0, "test"},
			out: "0.0",
		},
		{
			in:  configVersion{1, 0, "test"},
			out: "1.0",
		},
		{
			in:  configVersion{1, 1, "test"},
			out: "1.1",
		},
		{
			in:  configVersion{888, 66666, "test"},
			out: "888.66666",
		},
	}

	for _, tc := range cases {
		r := tc.in.ToString()
		assert.Equal(t, tc.out, r)
	}
}

// TestGetConfigVersion tests getting the version of a config file when
// the provided YAML is invalid.
func TestGetConfigVersion(t *testing.T) {
	cfg := ``

	_, err := getConfigVersion("test", []byte(cfg))
	assert.Error(t, err)
}

// TestGetConfigVersion2 tests getting the version of a config file when
// the config version is not supported.
func TestGetConfigVersion2(t *testing.T) {
	cfg := `version: "abc123 is not a supported version"`

	_, err := getConfigVersion("test", []byte(cfg))
	assert.Error(t, err)
}

// TestGetConfigVersion3 tests getting the version of a config file successfully.
func TestGetConfigVersion3(t *testing.T) {
	cfg := `version: 1.1`

	cv, err := getConfigVersion("test", []byte(cfg))
	assert.NoError(t, err)
	assert.Equal(t, 1, cv.Major)
	assert.Equal(t, 1, cv.Minor)
}

// TestGetConfigVersion4 tests getting the version of a config file unsuccessfully
// due to invalid YAML.
func TestGetConfigVersion4(t *testing.T) {
	cfg := `--~+::\n:-`

	cv, err := getConfigVersion("test", []byte(cfg))
	assert.Error(t, err)
	assert.Nil(t, cv)
}

// TestIsSupportedVersion tests checking if a config version is supported
// when the config version should not be supported.
func TestIsSupportedVersion(t *testing.T) {
	cv := configVersion{2, 0, "test"}

	isSupported := isSupportedVersion(&cv, []string{"1.0", "1.1"})
	assert.False(t, isSupported)
}

// TestIsSupportedVersion2 tests checking if a config version is supported
// when the config version should be supported.
func TestIsSupportedVersion2(t *testing.T) {
	cv := configVersion{1, 0, "test"}

	isSupported := isSupportedVersion(&cv, []string{"1.0", "1.1"})
	assert.True(t, isSupported)
}

// TestCfgVersionToConfigVersion tests converting a cfgVersion to
// a ConfigVersion when no version string is provided.
func TestCfgVersionToConfigVersion(t *testing.T) {
	c := cfgVersion{"", "test"}

	_, err := c.toConfigVersion()
	assert.Error(t, err)
}

// TestCfgVersionToConfigVersion2 tests converting a cfgVersion to
// a ConfigVersion when an invalid config value is provided.
func TestCfgVersionToConfigVersion2(t *testing.T) {
	c := cfgVersion{"abc", "test"}

	_, err := c.toConfigVersion()
	assert.Error(t, err)
}

// TestCfgVersionToConfigVersion3 tests converting a cfgVersion to
// a ConfigVersion when an invalid config value (major version) is supplied.
func TestCfgVersionToConfigVersion3(t *testing.T) {
	c := cfgVersion{"abc.0", "test"}

	_, err := c.toConfigVersion()
	assert.Error(t, err)
}

// TestCfgVersionToConfigVersion4 tests converting a cfgVersion to
// a ConfigVersion when an invalid config value (minor version) is supplied.
func TestCfgVersionToConfigVersion4(t *testing.T) {
	c := cfgVersion{"0.abc", "test"}

	_, err := c.toConfigVersion()
	assert.Error(t, err)
}

// TestCfgVersionToConfigVersion5 tests converting a cfgVersion to
// a ConfigVersion successfully, specifying only the major version.
func TestCfgVersionToConfigVersion5(t *testing.T) {
	c := cfgVersion{"1", "test"}

	cv, err := c.toConfigVersion()
	assert.NoError(t, err)

	expected := configVersion{1, 0, "test"}
	assert.Equal(t, expected, *cv)
}

// TestCfgVersionToConfigVersion6 tests converting a cfgVersion to
// a ConfigVersion successfully, specifying major and minor version.
func TestCfgVersionToConfigVersion6(t *testing.T) {
	c := cfgVersion{"1.1", "test"}

	cv, err := c.toConfigVersion()
	assert.NoError(t, err)

	expected := configVersion{1, 1, "test"}
	assert.Equal(t, expected, *cv)
}

// TestGetDeviceConfigVersionHandler tests getting the device config
// version handler for a version that is not supported.
func TestGetDeviceConfigVersionHandler(t *testing.T) {
	cv := configVersion{9999, 9999, "test"}

	_, err := getDeviceConfigVersionHandler(&cv)
	assert.Error(t, err)
}

// TestGetDeviceConfigVersionHandler2 tests getting the device config
// version handler for a version that is supported.
func TestGetDeviceConfigVersionHandler2(t *testing.T) {
	cv := configVersion{1, 0, "test"}

	h, err := getDeviceConfigVersionHandler(&cv)
	assert.NoError(t, err)
	assert.NotNil(t, h)
}

// TestGetDeviceConfigVersionHandler3 tests getting the device config
// version handler for a defined version (that is supported).
func TestGetDeviceConfigVersionHandler3(t *testing.T) {
	cv := v1maj0min

	h, err := getDeviceConfigVersionHandler(&cv)
	assert.NoError(t, err)
	assert.NotNil(t, h)
}

// TestGetDeviceConfigVersionHandler4 tests getting the device config
// version handler for a supported version when the handler does not exist.
func TestGetDeviceConfigVersionHandler4(t *testing.T) {
	cv := configVersion{999, 999, "test"}

	// add version 999.999 to the supported versions, but do not add a corresponding handler
	supportedDeviceConfigVersions = append(supportedDeviceConfigVersions, "999.999")

	h, err := getDeviceConfigVersionHandler(&cv)
	assert.Error(t, err)
	assert.Nil(t, h)
}

// TestGetPluginConfigVersionHandler tests getting the plugin config
// version handler for a version that is not supported.
func TestGetPluginConfigVersionHandler(t *testing.T) {
	cv := configVersion{9999, 9999, "test"}

	_, err := getPluginConfigVersionHandler(&cv)
	assert.Error(t, err)
}

// TestGetPluginConfigVersionHandler2 tests getting the plugin config
// version handler for a version that is supported.
func TestGetPluginConfigVersionHandler2(t *testing.T) {
	cv := configVersion{1, 0, "test"}

	h, err := getPluginConfigVersionHandler(&cv)
	assert.NoError(t, err)
	assert.NotNil(t, h)
}

// TestGetPluginConfigVersionHandler3 tests getting the plugin config
// version handler for a defined version (that is supported).
func TestGetPluginConfigVersionHandler3(t *testing.T) {
	cv := v1maj0min

	h, err := getPluginConfigVersionHandler(&cv)
	assert.NoError(t, err)
	assert.NotNil(t, h)
}

// TestGetPluginConfigVersionHandler4 tests getting the plugin config
// version handler for a supported version when the handler does not exist.
func TestGetPluginConfigVersionHandler4(t *testing.T) {
	cv := configVersion{999, 999, "test"}

	// add version 999.999 to the supported versions, but do not add a corresponding handler
	supportedPluginConfigVersions = append(supportedPluginConfigVersions, "999.999")

	h, err := getPluginConfigVersionHandler(&cv)
	assert.Error(t, err)
	assert.Nil(t, h)
}
