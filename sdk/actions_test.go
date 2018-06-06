package sdk

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// clearPreRunActions is a util to clear the pre run action slice.
func clearPreRunActions() {
	preRunActions = []pluginAction{}
}

// clearPostRunActions is a util to clear the post run action slice.
func clearPostRunActions() {
	postRunActions = []pluginAction{}
}

// clearDeviceSetupActions is a util to clear the device setup action map.
func clearDeviceSetupActions() {
	deviceSetupActions = map[string][]deviceAction{}
}

// TestActionsInit tests that the actions data structures were initialized
// via the init function.
func TestActionsInit(t *testing.T) {
	assert.NotNil(t, preRunActions)
	assert.Empty(t, preRunActions)
	assert.NotNil(t, postRunActions)
	assert.Empty(t, postRunActions)
	assert.NotNil(t, deviceSetupActions)
	assert.Empty(t, deviceSetupActions)
}

// Test_execPreRun tests running pre-run actions, when none are specified.
func Test_execPreRun(t *testing.T) {
	defer clearPreRunActions()

	plugin := NewPlugin()
	err := execPreRun(plugin)

	assert.NoError(t, err.Err())
}

// Test_execPreRun1 tests running pre-run actions, when one is specified.
func Test_execPreRun1(t *testing.T) {
	defer clearPreRunActions()

	c := 0
	action := func(_ *Plugin) error {
		c += 1
		return nil
	}

	plugin := NewPlugin()
	plugin.RegisterPreRunActions(action)

	err := execPreRun(plugin)

	assert.NoError(t, err.Err())
	assert.Equal(t, 1, c)
}

// Test_execPreRun2 tests running pre-run actions, when multiple are specified.
func Test_execPreRun2(t *testing.T) {
	defer clearPreRunActions()

	c := 0
	action := func(_ *Plugin) error {
		c += 1
		return nil
	}

	plugin := NewPlugin()
	plugin.RegisterPreRunActions(
		action,
		action,
		action,
	)

	err := execPreRun(plugin)

	assert.NoError(t, err.Err())
	assert.Equal(t, 3, c)
}

// Test_execPreRun3 tests running pre-run actions, when one is specified, but that
// one produces an error.
func Test_execPreRun3(t *testing.T) {
	defer clearPreRunActions()

	c := 0
	actionErr := func(_ *Plugin) error {
		return fmt.Errorf("error")
	}

	plugin := NewPlugin()
	plugin.RegisterPreRunActions(actionErr)

	err := execPreRun(plugin)

	assert.Error(t, err.Err())
	assert.Equal(t, 0, c)
}

// Test_execPreRun4 tests running pre-run actions, when multiple are specified, and
// one produces an error.
func Test_execPreRun4(t *testing.T) {
	defer clearPreRunActions()

	c := 0
	action := func(_ *Plugin) error {
		c += 1
		return nil
	}
	actionErr := func(_ *Plugin) error {
		return fmt.Errorf("error")
	}

	plugin := NewPlugin()
	plugin.RegisterPreRunActions(
		action,
		actionErr,
		action,
	)

	err := execPreRun(plugin)

	assert.Error(t, err.Err())
	assert.Equal(t, 2, c)
}

// Test_execPreRun5 tests running pre-run actions, when multiple are specified, and
// all produce an error.
func Test_execPreRun5(t *testing.T) {
	defer clearPreRunActions()

	c := 0
	actionErr := func(_ *Plugin) error {
		return fmt.Errorf("error")
	}

	plugin := NewPlugin()
	plugin.RegisterPreRunActions(
		actionErr,
		actionErr,
		actionErr,
	)

	err := execPreRun(plugin)

	assert.Error(t, err.Err())
	assert.Equal(t, 0, c)
}

// Test_execPostRun tests running post-run actions, when none are specified.
func Test_execPostRun(t *testing.T) {
	defer clearPostRunActions()

	plugin := NewPlugin()
	err := execPostRun(plugin)

	assert.NoError(t, err.Err())
}

// Test_execPostRun1 tests running post-run actions, when one is specified.
func Test_execPostRun1(t *testing.T) {
	defer clearPostRunActions()

	c := 0
	action := func(_ *Plugin) error {
		c += 1
		return nil
	}

	plugin := NewPlugin()
	plugin.RegisterPostRunActions(action)

	err := execPostRun(plugin)

	assert.NoError(t, err.Err())
	assert.Equal(t, 1, c)
}

// Test_execPostRun2 tests running post-run actions, when multiple are specified.
func Test_execPostRun2(t *testing.T) {
	defer clearPostRunActions()

	c := 0
	action := func(_ *Plugin) error {
		c += 1
		return nil
	}

	plugin := NewPlugin()
	plugin.RegisterPostRunActions(
		action,
		action,
		action,
	)

	err := execPostRun(plugin)

	assert.NoError(t, err.Err())
	assert.Equal(t, 3, c)
}

// Test_execPostRun3 tests running post-run actions, when one is specified, but
// that one produces an error.
func Test_execPostRun3(t *testing.T) {
	defer clearPostRunActions()

	c := 0
	actionErr := func(_ *Plugin) error {
		return fmt.Errorf("error")
	}

	plugin := NewPlugin()
	plugin.RegisterPostRunActions(actionErr)

	err := execPostRun(plugin)

	assert.Error(t, err.Err())
	assert.Equal(t, 0, c)
}

// Test_execPostRun4 tests running post-run actions, when multiple are specified,
// and one produces an error.
func Test_execPostRun4(t *testing.T) {
	defer clearPostRunActions()

	c := 0
	action := func(_ *Plugin) error {
		c += 1
		return nil
	}
	actionErr := func(_ *Plugin) error {
		return fmt.Errorf("error")
	}

	plugin := NewPlugin()
	plugin.RegisterPostRunActions(
		action,
		actionErr,
		action,
	)

	err := execPostRun(plugin)

	assert.Error(t, err.Err())
	assert.Equal(t, 2, c)
}

// Test_execPostRun5 tests running post-run actions, when multiple are specified,
// and all produce an error.
func Test_execPostRun5(t *testing.T) {
	defer clearPostRunActions()

	c := 0
	actionErr := func(_ *Plugin) error {
		return fmt.Errorf("error")
	}

	plugin := NewPlugin()
	plugin.RegisterPostRunActions(
		actionErr,
		actionErr,
		actionErr,
	)

	err := execPostRun(plugin)

	assert.Error(t, err.Err())
	assert.Equal(t, 0, c)
}

// Test_execDeviceSetup tests running device setup actions, when none are specified.
func Test_execDeviceSetup(t *testing.T) {
	defer clearDeviceSetupActions()

	plugin := NewPlugin()
	err := execDeviceSetup(plugin)

	assert.NoError(t, err.Err())
}

// Test_execDeviceSetup1 tests running device setup actions, when one exists, but the
// filter does not match anything.
func Test_execDeviceSetup1(t *testing.T) {
	defer clearDeviceSetupActions()

	c := 0
	action := func(_ *Plugin, _ *Device) error {
		c += 1
		return nil
	}

	plugin := NewPlugin()
	plugin.RegisterDeviceSetupActions("kind=foo", action)
	deviceMap["foobar"] = &Device{Kind: "test"}
	defer delete(deviceMap, "foobar")

	err := execDeviceSetup(plugin)

	assert.NoError(t, err.Err())
	assert.Equal(t, 0, c)
}

// Test_execDeviceSetup2 tests running device setup actions, when one exists and the
// filter matches a device.
func Test_execDeviceSetup2(t *testing.T) {
	defer clearDeviceSetupActions()

	c := 0
	action := func(_ *Plugin, _ *Device) error {
		c += 1
		return nil
	}

	plugin := NewPlugin()
	plugin.RegisterDeviceSetupActions("kind=test", action)
	deviceMap["foobar"] = &Device{Kind: "test"}
	defer delete(deviceMap, "foobar")

	err := execDeviceSetup(plugin)

	assert.NoError(t, err.Err())
	assert.Equal(t, 1, c)
}

// Test_execDeviceSetup3 tests running device setup actions, when an invalid filter
// is specified.
func Test_execDeviceSetup3(t *testing.T) {
	defer clearDeviceSetupActions()

	c := 0
	action := func(_ *Plugin, _ *Device) error {
		c += 1
		return nil
	}

	plugin := NewPlugin()
	plugin.RegisterDeviceSetupActions("foo=test", action)
	deviceMap["foobar"] = &Device{Kind: "test"}
	defer delete(deviceMap, "foobar")

	err := execDeviceSetup(plugin)

	assert.Error(t, err.Err())
	assert.Equal(t, 0, c)
}

// Test_execDeviceSetup4 tests running device setup actions, when there are multiple
// actions for a device.
func Test_execDeviceSetup4(t *testing.T) {
	defer clearDeviceSetupActions()

	c := 0
	action := func(_ *Plugin, _ *Device) error {
		c += 1
		return nil
	}

	plugin := NewPlugin()
	plugin.RegisterDeviceSetupActions("kind=test", action, action, action)
	deviceMap["foobar"] = &Device{Kind: "test"}
	defer delete(deviceMap, "foobar")

	err := execDeviceSetup(plugin)

	assert.NoError(t, err.Err())
	assert.Equal(t, 3, c)

}

// Test_execDeviceSetup5 tests running device setup actions, when a single action exists
// and fails.
func Test_execDeviceSetup5(t *testing.T) {
	defer clearDeviceSetupActions()

	c := 0
	actionErr := func(_ *Plugin, _ *Device) error {
		return fmt.Errorf("error")
	}

	plugin := NewPlugin()
	plugin.RegisterDeviceSetupActions("kind=test", actionErr)
	deviceMap["foobar"] = &Device{Kind: "test"}
	defer delete(deviceMap, "foobar")

	err := execDeviceSetup(plugin)

	assert.Error(t, err.Err())
	assert.Equal(t, 0, c)
}

// Test_execDeviceSetup6 tests running device setup actions, when multiple actions exist
// and some fail.
func Test_execDeviceSetup6(t *testing.T) {
	defer clearDeviceSetupActions()

	c := 0
	action := func(_ *Plugin, _ *Device) error {
		c += 1
		return nil
	}
	actionErr := func(_ *Plugin, _ *Device) error {
		return fmt.Errorf("error")
	}

	plugin := NewPlugin()
	plugin.RegisterDeviceSetupActions("kind=test", action, actionErr, action)
	deviceMap["foobar"] = &Device{Kind: "test"}
	defer delete(deviceMap, "foobar")

	err := execDeviceSetup(plugin)

	assert.Error(t, err.Err())
	assert.Equal(t, 2, c)
}

// Test_execDeviceSetup7 tests running device setup actions, when multiple actions exist
// and all fail.
func Test_execDeviceSetup7(t *testing.T) {
	defer clearDeviceSetupActions()

	c := 0
	actionErr := func(_ *Plugin, _ *Device) error {
		return fmt.Errorf("error")
	}

	plugin := NewPlugin()
	plugin.RegisterDeviceSetupActions("kind=test", actionErr, actionErr, actionErr)
	deviceMap["foobar"] = &Device{Kind: "test"}
	defer delete(deviceMap, "foobar")

	err := execDeviceSetup(plugin)

	assert.Error(t, err.Err())
	assert.Equal(t, 0, c)
}
