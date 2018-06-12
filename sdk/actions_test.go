package sdk

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test_execPreRun tests running pre-run actions, when none are specified.
func Test_execPreRun(t *testing.T) {
	defer resetContext()

	plugin := NewPlugin()
	err := execPreRun(plugin)

	assert.NoError(t, err.Err())
}

// Test_execPreRun1 tests running pre-run actions, when one is specified.
func Test_execPreRun1(t *testing.T) {
	defer resetContext()

	c := 0
	action := func(_ *Plugin) error {
		c++
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
	defer resetContext()

	c := 0
	action := func(_ *Plugin) error {
		c++
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
	defer resetContext()

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
	defer resetContext()

	c := 0
	action := func(_ *Plugin) error {
		c++
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
	defer resetContext()

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
	defer resetContext()

	plugin := NewPlugin()
	err := execPostRun(plugin)

	assert.NoError(t, err.Err())
}

// Test_execPostRun1 tests running post-run actions, when one is specified.
func Test_execPostRun1(t *testing.T) {
	defer resetContext()

	c := 0
	action := func(_ *Plugin) error {
		c++
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
	defer resetContext()

	c := 0
	action := func(_ *Plugin) error {
		c++
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
	defer resetContext()

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
	defer resetContext()

	c := 0
	action := func(_ *Plugin) error {
		c++
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
	defer resetContext()

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
	defer resetContext()

	plugin := NewPlugin()
	err := execDeviceSetup(plugin)

	assert.NoError(t, err.Err())
}

// Test_execDeviceSetup1 tests running device setup actions, when one exists, but the
// filter does not match anything.
func Test_execDeviceSetup1(t *testing.T) {
	defer resetContext()

	c := 0
	action := func(_ *Plugin, _ *Device) error {
		c++
		return nil
	}

	plugin := NewPlugin()
	plugin.RegisterDeviceSetupActions("kind=foo", action)
	ctx.devices["foobar"] = &Device{Kind: "test"}

	err := execDeviceSetup(plugin)

	assert.NoError(t, err.Err())
	assert.Equal(t, 0, c)
}

// Test_execDeviceSetup2 tests running device setup actions, when one exists and the
// filter matches a device.
func Test_execDeviceSetup2(t *testing.T) {
	defer resetContext()

	c := 0
	action := func(_ *Plugin, _ *Device) error {
		c++
		return nil
	}

	plugin := NewPlugin()
	plugin.RegisterDeviceSetupActions("kind=test", action)
	ctx.devices["foobar"] = &Device{Kind: "test"}

	err := execDeviceSetup(plugin)

	assert.NoError(t, err.Err())
	assert.Equal(t, 1, c)
}

// Test_execDeviceSetup3 tests running device setup actions, when an invalid filter
// is specified.
func Test_execDeviceSetup3(t *testing.T) {
	defer resetContext()

	c := 0
	action := func(_ *Plugin, _ *Device) error {
		c++
		return nil
	}

	plugin := NewPlugin()
	plugin.RegisterDeviceSetupActions("foo=test", action)
	ctx.devices["foobar"] = &Device{Kind: "test"}

	err := execDeviceSetup(plugin)

	assert.Error(t, err.Err())
	assert.Equal(t, 0, c)
}

// Test_execDeviceSetup4 tests running device setup actions, when there are multiple
// actions for a device.
func Test_execDeviceSetup4(t *testing.T) {
	defer resetContext()

	c := 0
	action := func(_ *Plugin, _ *Device) error {
		c++
		return nil
	}

	plugin := NewPlugin()
	plugin.RegisterDeviceSetupActions("kind=test", action, action, action)
	ctx.devices["foobar"] = &Device{Kind: "test"}

	err := execDeviceSetup(plugin)

	assert.NoError(t, err.Err())
	assert.Equal(t, 3, c)

}

// Test_execDeviceSetup5 tests running device setup actions, when a single action exists
// and fails.
func Test_execDeviceSetup5(t *testing.T) {
	defer resetContext()

	c := 0
	actionErr := func(_ *Plugin, _ *Device) error {
		return fmt.Errorf("error")
	}

	plugin := NewPlugin()
	plugin.RegisterDeviceSetupActions("kind=test", actionErr)
	ctx.devices["foobar"] = &Device{Kind: "test"}

	err := execDeviceSetup(plugin)

	assert.Error(t, err.Err())
	assert.Equal(t, 0, c)
}

// Test_execDeviceSetup6 tests running device setup actions, when multiple actions exist
// and some fail.
func Test_execDeviceSetup6(t *testing.T) {
	defer resetContext()

	c := 0
	action := func(_ *Plugin, _ *Device) error {
		c++
		return nil
	}
	actionErr := func(_ *Plugin, _ *Device) error {
		return fmt.Errorf("error")
	}

	plugin := NewPlugin()
	plugin.RegisterDeviceSetupActions("kind=test", action, actionErr, action)
	ctx.devices["foobar"] = &Device{Kind: "test"}

	err := execDeviceSetup(plugin)

	assert.Error(t, err.Err())
	assert.Equal(t, 2, c)
}

// Test_execDeviceSetup7 tests running device setup actions, when multiple actions exist
// and all fail.
func Test_execDeviceSetup7(t *testing.T) {
	defer resetContext()

	c := 0
	actionErr := func(_ *Plugin, _ *Device) error {
		return fmt.Errorf("error")
	}

	plugin := NewPlugin()
	plugin.RegisterDeviceSetupActions("kind=test", actionErr, actionErr, actionErr)
	ctx.devices["foobar"] = &Device{Kind: "test"}

	err := execDeviceSetup(plugin)

	assert.Error(t, err.Err())
	assert.Equal(t, 0, c)
}
