package config

import (
	"testing"
)

var configVersionToStringTestTable = []struct {
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

func TestConfigVersion_ToString(t *testing.T) {
	for _, tc := range configVersionToStringTestTable {
		r := tc.in.ToString()
		if r != tc.out {
			t.Errorf("%#v.ToString() => %v, want %v", tc.in, r, tc.out)
		}
	}
}

func TestGetConfigVersion(t *testing.T) {
	cfg := ``

	_, err := getConfigVersion("test", []byte(cfg))
	if err == nil {
		t.Error("expected error: given YAML should be invalid")
	}
}

func TestGetConfigVersion2(t *testing.T) {
	cfg := `version: "abc123 is not a supported version"`

	_, err := getConfigVersion("test", []byte(cfg))
	if err == nil {
		t.Error("expected error: given version configuration is invalid")
	}
}

func TestGetConfigVersion3(t *testing.T) {
	cfg := `version: 1.1`

	cv, err := getConfigVersion("test", []byte(cfg))
	if err != nil {
		t.Error(err)
	}
	if cv.Major != 1 {
		t.Errorf("expected parsed major version to be 1, but was %v", cv.Major)
	}
	if cv.Minor != 1 {
		t.Errorf("expected parsed minor version to be 1, but was %v", cv.Minor)
	}
}

func TestIsSupportedVersion(t *testing.T) {
	cv := configVersion{2, 0, "test"}

	isSupported := isSupportedVersion(&cv, []string{"1.0", "1.1"})
	if isSupported {
		t.Error("expected config version to fail supported check")
	}
}

func TestIsSupportedVersion2(t *testing.T) {
	cv := configVersion{1, 0, "test"}

	isSupported := isSupportedVersion(&cv, []string{"1.0", "1.1"})
	if !isSupported {
		t.Error("expected config version to pass supported check")
	}
}

func TestCfgVersionToConfigVersion(t *testing.T) {
	c := cfgVersion{"", "test"}

	_, err := c.toConfigVersion()
	if err == nil {
		t.Error("expected error: no version string provided")
	}
}

func TestCfgVersionToConfigVersion2(t *testing.T) {
	c := cfgVersion{"abc", "test"}

	_, err := c.toConfigVersion()
	if err == nil {
		t.Error("expected error: invalid config value")
	}
}

func TestCfgVersionToConfigVersion3(t *testing.T) {
	c := cfgVersion{"abc.0", "test"}

	_, err := c.toConfigVersion()
	if err == nil {
		t.Error("expected error: invalid config value (major version)")
	}
}

func TestCfgVersionToConfigVersion4(t *testing.T) {
	c := cfgVersion{"0.abc", "test"}

	_, err := c.toConfigVersion()
	if err == nil {
		t.Error("expected error: invalid config value (minor version)")
	}
}

func TestCfgVersionToConfigVersion5(t *testing.T) {
	c := cfgVersion{"1", "test"}

	cv, err := c.toConfigVersion()
	if err != nil {
		t.Error(err)
	}

	expected := configVersion{1, 0, "test"}
	if *cv != expected {
		t.Errorf("expected version to match 1.0, but instead is %v", cv)
	}
}

func TestCfgVersionToConfigVersion6(t *testing.T) {
	c := cfgVersion{"1.1", "test"}

	cv, err := c.toConfigVersion()
	if err != nil {
		t.Error(err)
	}

	expected := configVersion{1, 1, "test"}
	if *cv != expected {
		t.Errorf("expected version to match 1.1, but instead is %v", cv)
	}
}

func TestGetDeviceConfigVersionHandler(t *testing.T) {
	cv := configVersion{9999, 9999, "test"}

	_, err := getDeviceConfigVersionHandler(&cv)
	if err == nil {
		t.Error("expected error: config version should not be supported")
	}
}

func TestGetDeviceConfigVersionHandler2(t *testing.T) {
	cv := configVersion{1, 0, "test"}

	h, err := getDeviceConfigVersionHandler(&cv)
	if err != nil {
		t.Error(err)
	}
	if h == nil {
		t.Error("no handler returned, but one was expected")
	}
}

func TestGetDeviceConfigVersionHandler3(t *testing.T) {
	cv := v1maj0min

	h, err := getDeviceConfigVersionHandler(&cv)
	if err != nil {
		t.Error(err)
	}
	if h == nil {
		t.Error("no handler returned, but one was expected")
	}
}

func TestGetPluginConfigVersionHandler(t *testing.T) {
	cv := configVersion{9999, 9999, "test"}

	_, err := getPluginConfigVersionHandler(&cv)
	if err == nil {
		t.Error("expected error: config version should not be supported")
	}
}

func TestGetPluginConfigVersionHandler2(t *testing.T) {
	cv := configVersion{1, 0, "test"}

	h, err := getPluginConfigVersionHandler(&cv)
	if err != nil {
		t.Error(err)
	}
	if h == nil {
		t.Error("no handler returned, but one was expected")
	}
}

func TestGetPluginConfigVersionHandler3(t *testing.T) {
	cv := v1maj0min

	h, err := getPluginConfigVersionHandler(&cv)
	if err != nil {
		t.Error(err)
	}
	if h == nil {
		t.Error("no handler returned, but one was expected")
	}
}
