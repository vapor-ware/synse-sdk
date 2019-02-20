package sdk

//
//import (
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//	"github.com/vapor-ware/synse-sdk/sdk/errors"
//)
//
//// TestNewPlugin tests creating a new plugin.
//func TestNewPlugin(t *testing.T) {
//	var testTable = []struct {
//		desc    string
//		options []PluginOption
//	}{
//		{
//			desc:    "no plugin options",
//			options: []PluginOption{},
//		},
//		{
//			desc: "one options",
//			options: []PluginOption{
//				CustomDeviceIdentifier(func(i map[string]interface{}) string {
//					return "test"
//				}),
//			},
//		},
//		{
//			desc: "multiple options",
//			options: []PluginOption{
//				CustomDeviceIdentifier(func(i map[string]interface{}) string {
//					return "test"
//				}),
//				CustomDynamicDeviceRegistration(func(i map[string]interface{}) ([]*Device, error) {
//					return nil, nil
//				}),
//			},
//		},
//		{
//			desc: "duplicate options",
//			options: []PluginOption{
//				CustomDeviceIdentifier(func(i map[string]interface{}) string {
//					return "foo"
//				}),
//				CustomDeviceIdentifier(func(i map[string]interface{}) string {
//					return "bar"
//				}),
//			},
//		},
//	}
//
//	for _, testCase := range testTable {
//		plugin := NewPlugin(testCase.options...)
//		assert.NotNil(t, plugin)
//	}
//}
//
//// TestPlugin_RegisterOutputTypes tests registering the output types for the plugin.
//func TestPlugin_RegisterOutputTypes(t *testing.T) {
//	defer resetContext()
//
//	types := []*OutputType{
//		{Name: "foo"},
//		{Name: "bar"},
//		{Name: "baz"},
//	}
//	plugin := NewPlugin()
//	assert.Equal(t, 0, len(ctx.outputTypes))
//
//	err := plugin.RegisterOutputTypes(types...)
//	assert.NoError(t, err)
//	assert.Equal(t, 3, len(ctx.outputTypes))
//}
//
//// TestPlugin_RegisterOutputTypesError tests registering the output types for the
//// plugin when duplicate types are specified.
//func TestPlugin_RegisterOutputTypesError(t *testing.T) {
//	defer resetContext()
//
//	types := []*OutputType{
//		{Name: "foo"},
//		{Name: "bar"},
//		{Name: "foo"},
//	}
//	plugin := NewPlugin()
//	assert.Equal(t, 0, len(ctx.outputTypes))
//
//	err := plugin.RegisterOutputTypes(types...)
//	assert.Error(t, err)
//	assert.True(t, len(ctx.outputTypes) > 0)
//}
//
//// TestPlugin_RegisterPreRunActions tests registering pre-run actions.
//func TestPlugin_RegisterPreRunActions(t *testing.T) {
//	defer resetContext()
//
//	actions := []pluginAction{
//		func(_ *Plugin) error { return nil },
//		func(_ *Plugin) error { return nil },
//		func(_ *Plugin) error { return nil },
//	}
//
//	plugin := NewPlugin()
//
//	assert.Equal(t, 0, len(ctx.preRunActions))
//	plugin.RegisterPreRunActions(actions...)
//	assert.Equal(t, 3, len(ctx.preRunActions))
//}
//
//// TestPlugin_RegisterPostRunActions tests registering post-run actions.
//func TestPlugin_RegisterPostRunActions(t *testing.T) {
//	defer resetContext()
//
//	actions := []pluginAction{
//		func(_ *Plugin) error { return nil },
//		func(_ *Plugin) error { return nil },
//		func(_ *Plugin) error { return nil },
//	}
//
//	plugin := NewPlugin()
//
//	assert.Equal(t, 0, len(ctx.postRunActions))
//	plugin.RegisterPostRunActions(actions...)
//	assert.Equal(t, 3, len(ctx.postRunActions))
//}
//
//// TestPlugin_RegisterDeviceSetupActions tests registering device setup actions.
//func TestPlugin_RegisterDeviceSetupActions(t *testing.T) {
//	defer resetContext()
//
//	action := func(_ *Plugin, _ *Device) error { return nil }
//
//	plugin := NewPlugin()
//
//	assert.Equal(t, 0, len(ctx.deviceSetupActions))
//	plugin.RegisterDeviceSetupActions("kind=test", action, action, action)
//	assert.Equal(t, 1, len(ctx.deviceSetupActions))
//	assert.Equal(t, 3, len(ctx.deviceSetupActions["kind=test"]))
//}
//
//// TestPlugin_RegisterDeviceSetupActions2 tests registering device setup actions when
//// some already exist.
//func TestPlugin_RegisterDeviceSetupActions2(t *testing.T) {
//	defer resetContext()
//
//	action := func(_ *Plugin, _ *Device) error { return nil }
//
//	// add something to the device setup actions to start with
//	ctx.deviceSetupActions["kind=test"] = []deviceAction{action}
//
//	plugin := NewPlugin()
//
//	assert.Equal(t, 1, len(ctx.deviceSetupActions))
//	plugin.RegisterDeviceSetupActions("kind=test", action)
//	assert.Equal(t, 1, len(ctx.deviceSetupActions))
//	assert.Equal(t, 2, len(ctx.deviceSetupActions["kind=test"]))
//}
//
//// TestPlugin_RegisterDeviceHandlers tests registering DeviceHandlers with the plugin.
//func TestPlugin_RegisterDeviceHandlers(t *testing.T) {
//	defer resetContext()
//
//	fooHandler := &DeviceHandler{Name: "foo"}
//	barHandler := &DeviceHandler{Name: "bar"}
//
//	plugin := NewPlugin()
//
//	assert.Equal(t, 0, len(ctx.deviceHandlers))
//	plugin.RegisterDeviceHandlers(fooHandler, barHandler)
//	assert.Equal(t, 2, len(ctx.deviceHandlers))
//}
//
//// TestNewDefaultPluginConfig tests getting a default plugin config.
//func TestNewDefaultPluginConfig(t *testing.T) {
//	cfg, err := NewDefaultPluginConfig()
//	assert.NoError(t, err)
//	assert.NotNil(t, cfg)
//
//	assert.Equal(t, false, cfg.Debug)
//	assert.NotNil(t, cfg.Settings, "settings should not be nil")
//	assert.NotNil(t, cfg.Network, "network should not be nil")
//	assert.NotNil(t, cfg.DynamicRegistration, "dynamic registration should not be nil")
//	assert.NotNil(t, cfg.Context, "context should not be nil")
//
//	assert.Nil(t, cfg.Limiter, "limiter should be nil")
//}
//
//// TestPluginConfig_Validate_Ok tests validating a PluginConfig with no errors.
//func TestPluginConfig_Validate_Ok(t *testing.T) {
//	var testTable = []struct {
//		desc   string
//		config PluginConfig
//	}{
//		{
//			desc: "PluginConfig has valid version and network",
//			config: PluginConfig{
//				Version: 1,
//				Network: &NetworkSettings{
//					Type:    "tcp",
//					Address: "10.10.10.10",
//				},
//			},
//		},
//		{
//			desc: "PluginConfig has valid version and network, invalid settings (not validated here)",
//			config: PluginConfig{
//				Version: 1,
//				Network: &NetworkSettings{
//					Type:    "tcp",
//					Address: "10.10.10.10",
//				},
//				Settings: &PluginSettings{
//					Mode: "bad mode",
//				},
//			},
//		},
//	}
//
//	for _, testCase := range testTable {
//		merr := errors.NewMultiError("test")
//
//		testCase.config.Validate(merr)
//		assert.NoError(t, merr.Err(), testCase.desc)
//	}
//}
//
//// TestPluginConfig_Validate_Error tests validating a PluginConfig with errors.
//func TestPluginConfig_Validate_Error(t *testing.T) {
//	var testTable = []struct {
//		desc     string
//		errCount int
//		config   PluginConfig
//	}{
//		{
//			desc:     "PluginConfig has no network",
//			errCount: 1,
//			config: PluginConfig{
//				Version: 1,
//			},
//		},
//		{
//			desc:     "PluginConfig is empty",
//			errCount: 1,
//			config:   PluginConfig{},
//		},
//	}
//
//	for _, testCase := range testTable {
//		merr := errors.NewMultiError("test")
//
//		testCase.config.Validate(merr)
//		assert.Error(t, merr.Err(), testCase.desc)
//		assert.Equal(t, testCase.errCount, len(merr.Errors), merr.Error())
//	}
//}
//
//// TestPluginSettings_Validate_Ok tests validating a PluginSettings with no errors.
//func TestPluginSettings_Validate_Ok(t *testing.T) {
//	var testTable = []struct {
//		desc   string
//		config PluginSettings
//	}{
//		{
//			desc: "PluginSettings has valid mode (serial)",
//			config: PluginSettings{
//				Mode:        "serial",
//				Read:        &ReadSettings{},
//				Write:       &WriteSettings{},
//				Transaction: &TransactionSettings{},
//			},
//		},
//		{
//			desc: "PluginSettings has valid mode (parallel)",
//			config: PluginSettings{
//				Mode:        "parallel",
//				Read:        &ReadSettings{},
//				Write:       &WriteSettings{},
//				Transaction: &TransactionSettings{},
//			},
//		},
//	}
//
//	for _, testCase := range testTable {
//		merr := errors.NewMultiError("test")
//
//		testCase.config.Validate(merr)
//		assert.NoError(t, merr.Err(), testCase.desc)
//	}
//}
//
//// TestPluginSettings_Validate_Error tests validating a PluginSettings with errors.
//func TestPluginSettings_Validate_Error(t *testing.T) {
//	var testTable = []struct {
//		desc     string
//		errCount int
//		config   PluginSettings
//	}{
//		{
//			desc:     "PluginSettings is empty",
//			errCount: 1,
//			config:   PluginSettings{},
//		},
//		{
//			desc:     "PluginSettings has invalid mode",
//			errCount: 1,
//			config: PluginSettings{
//				Mode:        "bad mode",
//				Read:        &ReadSettings{},
//				Write:       &WriteSettings{},
//				Transaction: &TransactionSettings{},
//			},
//		},
//	}
//
//	for _, testCase := range testTable {
//		merr := errors.NewMultiError("test")
//
//		testCase.config.Validate(merr)
//		assert.Error(t, merr.Err(), testCase.desc)
//		assert.Equal(t, testCase.errCount, len(merr.Errors), merr.Error())
//	}
//}
//
//// TestPluginSettings_IsParallel tests if the plugin is in parallel mode.
//func TestPluginSettings_IsParallel(t *testing.T) {
//	parallel := PluginSettings{Mode: "parallel"}
//	assert.True(t, parallel.IsParallel())
//
//	serial := PluginSettings{Mode: "serial"}
//	assert.False(t, serial.IsParallel())
//}
//
//// TestPluginSettings_IsSerial tests if the plugin is in serial mode.
//func TestPluginSettings_IsSerial(t *testing.T) {
//	parallel := PluginSettings{Mode: "parallel"}
//	assert.False(t, parallel.IsSerial())
//
//	serial := PluginSettings{Mode: "serial"}
//	assert.True(t, serial.IsSerial())
//}
//
//// TestNetworkSettings_Validate_Ok tests validating a NetworkSettings with no errors.
//func TestNetworkSettings_Validate_Ok(t *testing.T) {
//	var testTable = []struct {
//		desc   string
//		config NetworkSettings
//	}{
//		{
//			desc: "NetworkSettings has valid type (tcp) and address",
//			config: NetworkSettings{
//				Type:    "tcp",
//				Address: "1.2.3.4",
//			},
//		},
//		{
//			desc: "NetworkSettings has valid type (unix) and address",
//			config: NetworkSettings{
//				Type:    "unix",
//				Address: "foo.sock",
//			},
//		},
//	}
//
//	for _, testCase := range testTable {
//		merr := errors.NewMultiError("test")
//
//		testCase.config.Validate(merr)
//		assert.NoError(t, merr.Err(), testCase.desc)
//	}
//}
//
//// TestNetworkSettings_Validate_Error tests validating a NetworkSettings with errors.
//func TestNetworkSettings_Validate_Error(t *testing.T) {
//	var testTable = []struct {
//		desc     string
//		errCount int
//		config   NetworkSettings
//	}{
//		{
//			desc:     "NetworkSettings is empty",
//			errCount: 2,
//			config:   NetworkSettings{},
//		},
//		{
//			desc:     "NetworkSettings has no type",
//			errCount: 1,
//			config: NetworkSettings{
//				Address: "1.2.3.4",
//			},
//		},
//		{
//			desc:     "NetworkSettings has no address",
//			errCount: 1,
//			config: NetworkSettings{
//				Type: "tcp",
//			},
//		},
//		{
//			desc:     "NetworkSettings has invalid type",
//			errCount: 1,
//			config: NetworkSettings{
//				Type:    "other",
//				Address: "foo",
//			},
//		},
//	}
//
//	for _, testCase := range testTable {
//		merr := errors.NewMultiError("test")
//
//		testCase.config.Validate(merr)
//		assert.Error(t, merr.Err(), testCase.desc)
//		assert.Equal(t, testCase.errCount, len(merr.Errors), merr.Error())
//	}
//}
//
//// TestLimiterSettings_Validate_Ok tests validating a LimiterSettings with no errors.
//func TestLimiterSettings_Validate_Ok(t *testing.T) {
//	var testTable = []struct {
//		desc   string
//		config LimiterSettings
//	}{
//		{
//			desc:   "LimiterSettings is empty",
//			config: LimiterSettings{},
//		},
//		{
//			desc: "LimiterSettings has valid rate (0)",
//			config: LimiterSettings{
//				Rate: 0,
//			},
//		},
//		{
//			desc: "LimiterSettings has valid burst (0)",
//			config: LimiterSettings{
//				Burst: 0,
//			},
//		},
//		{
//			desc: "LimiterSettings has valid rate (>0)",
//			config: LimiterSettings{
//				Rate: 100,
//			},
//		},
//		{
//			desc: "LimiterSettings has valid burst (>0)",
//			config: LimiterSettings{
//				Burst: 100,
//			},
//		},
//		{
//			desc: "LimiterSettings has valid rate and burst",
//			config: LimiterSettings{
//				Rate:  100,
//				Burst: 100,
//			},
//		},
//	}
//
//	for _, testCase := range testTable {
//		merr := errors.NewMultiError("test")
//
//		testCase.config.Validate(merr)
//		assert.NoError(t, merr.Err(), testCase.desc)
//	}
//}
//
//// TestLimiterSettings_Validate_Error tests validating a LimiterSettings with errors.
//func TestLimiterSettings_Validate_Error(t *testing.T) {
//	var testTable = []struct {
//		desc     string
//		errCount int
//		config   LimiterSettings
//	}{
//		{
//			desc:     "LimiterSettings has rate below 0",
//			errCount: 1,
//			config: LimiterSettings{
//				Rate: -1,
//			},
//		},
//		{
//			desc:     "LimiterSettings has burst below 0",
//			errCount: 1,
//			config: LimiterSettings{
//				Burst: -1,
//			},
//		},
//		{
//			desc:     "LimiterSettings has rate and burst below 0",
//			errCount: 2,
//			config: LimiterSettings{
//				Rate:  -1,
//				Burst: -1,
//			},
//		},
//	}
//
//	for _, testCase := range testTable {
//		merr := errors.NewMultiError("test")
//
//		testCase.config.Validate(merr)
//		assert.Error(t, merr.Err(), testCase.desc)
//		assert.Equal(t, testCase.errCount, len(merr.Errors), merr.Error())
//	}
//}
//
//// TestReadSettings_Validate_Ok tests validating a ReadSettings with no errors.
//func TestReadSettings_Validate_Ok(t *testing.T) {
//	var testTable = []struct {
//		desc   string
//		config ReadSettings
//	}{
//		{
//			desc: "ReadSettings has valid interval and buffer size and SerialReadInterval",
//			config: ReadSettings{
//				Interval:           "5s",
//				Buffer:             100,
//				SerialReadInterval: "0s",
//			},
//		},
//	}
//
//	for _, testCase := range testTable {
//		merr := errors.NewMultiError("test")
//
//		testCase.config.Validate(merr)
//		assert.NoError(t, merr.Err(), testCase.desc)
//	}
//}
//
//// TestReadSettings_Validate_Error tests validating a ReadSettings with errors.
//func TestReadSettings_Validate_Error(t *testing.T) {
//	var testTable = []struct {
//		desc     string
//		errCount int
//		config   ReadSettings
//	}{
//		{
//			desc:     "ReadSettings has invalid interval",
//			errCount: 1,
//			config: ReadSettings{
//				Interval:           "foobar",
//				Buffer:             100,
//				SerialReadInterval: "1s",
//			},
//		},
//		{
//			desc:     "ReadSettings has invalid serial read interval",
//			errCount: 1,
//			config: ReadSettings{
//				Interval:           "5s",
//				Buffer:             100,
//				SerialReadInterval: "invalid",
//			},
//		},
//		{
//			desc:     "ReadSettings has invalid buffer size",
//			errCount: 1,
//			config: ReadSettings{
//				Interval:           "1s",
//				Buffer:             0,
//				SerialReadInterval: "1s",
//			},
//		},
//		{
//			desc:     "ReadSettings has invalid interval and invalid buffer size",
//			errCount: 2,
//			config: ReadSettings{
//				Interval:           "xyz",
//				Buffer:             -1,
//				SerialReadInterval: "1s",
//			},
//		},
//		{
//			desc:     "ReadSettings is empty",
//			errCount: 3,
//			config:   ReadSettings{},
//		},
//	}
//
//	for _, testCase := range testTable {
//		merr := errors.NewMultiError("test")
//
//		testCase.config.Validate(merr)
//		assert.Error(t, merr.Err(), testCase.desc)
//		assert.Equal(t, testCase.errCount, len(merr.Errors), merr.Error())
//	}
//}
//
//// TestWriteSettings_Validate_Ok tests validating a WriteSettings with no errors.
//func TestWriteSettings_Validate_Ok(t *testing.T) {
//	var testTable = []struct {
//		desc   string
//		config WriteSettings
//	}{
//		{
//			desc: "WriteSettings has valid interval, buffer, and max",
//			config: WriteSettings{
//				Interval: "5s",
//				Buffer:   100,
//				Max:      100,
//			},
//		},
//	}
//
//	for _, testCase := range testTable {
//		merr := errors.NewMultiError("test")
//
//		testCase.config.Validate(merr)
//		assert.NoError(t, merr.Err(), testCase.desc)
//	}
//}
//
//// TestWriteSettings_Validate_Error tests validating a WriteSettings with errors.
//func TestWriteSettings_Validate_Error(t *testing.T) {
//	var testTable = []struct {
//		desc     string
//		errCount int
//		config   WriteSettings
//	}{
//		{
//			desc:     "WriteSettings has invalid interval",
//			errCount: 1,
//			config: WriteSettings{
//				Interval: "foobar",
//				Buffer:   100,
//				Max:      100,
//			},
//		},
//		{
//			desc:     "WriteSettings has invalid buffer",
//			errCount: 1,
//			config: WriteSettings{
//				Interval: "5s",
//				Buffer:   0,
//				Max:      100,
//			},
//		},
//		{
//			desc:     "WriteSettings has invalid max",
//			errCount: 1,
//			config: WriteSettings{
//				Interval: "5s",
//				Buffer:   100,
//				Max:      0,
//			},
//		},
//		{
//			desc:     "WriteSettings has invalid interval, buffer, and max",
//			errCount: 3,
//			config: WriteSettings{
//				Interval: "xyz",
//				Buffer:   -1,
//				Max:      -100,
//			},
//		},
//		{
//			desc:     "WriteSettings is empty",
//			errCount: 3,
//			config:   WriteSettings{},
//		},
//	}
//
//	for _, testCase := range testTable {
//		merr := errors.NewMultiError("test")
//
//		testCase.config.Validate(merr)
//		assert.Error(t, merr.Err(), testCase.desc)
//		assert.Equal(t, testCase.errCount, len(merr.Errors), merr.Error())
//	}
//}
//
//// TestTransactionSettings_Validate_Ok tests validating a TransactionSettings with no errors.
//func TestTransactionSettings_Validate_Ok(t *testing.T) {
//	var testTable = []struct {
//		desc   string
//		config TransactionSettings
//	}{
//		{
//			desc: "TransactionSettings has valid TTL",
//			config: TransactionSettings{
//				TTL: "5s",
//			},
//		},
//	}
//
//	for _, testCase := range testTable {
//		merr := errors.NewMultiError("test")
//
//		testCase.config.Validate(merr)
//		assert.NoError(t, merr.Err(), testCase.desc)
//	}
//}
//
//// TestTransactionSettings_Validate_Error tests validating a TransactionSettings with errors.
//func TestTransactionSettings_Validate_Error(t *testing.T) {
//	var testTable = []struct {
//		desc     string
//		errCount int
//		config   TransactionSettings
//	}{
//		{
//			desc:     "TransactionSettings is empty",
//			errCount: 1,
//			config:   TransactionSettings{},
//		},
//		{
//			desc:     "TransactionSettings has invalid TTL",
//			errCount: 1,
//			config: TransactionSettings{
//				TTL: "xyz",
//			},
//		},
//	}
//
//	for _, testCase := range testTable {
//		merr := errors.NewMultiError("test")
//
//		testCase.config.Validate(merr)
//		assert.Error(t, merr.Err(), testCase.desc)
//		assert.Equal(t, testCase.errCount, len(merr.Errors), merr.Error())
//	}
//}
//
//// TestDynamicRegistrationSettings_Validate tests validating a DynamicRegistrationSettings.
//// Validation should always pass here.
//func TestDynamicRegistrationSettings_Validate(t *testing.T) {
//	merr := errors.NewMultiError("test")
//	config := DynamicRegistrationSettings{}
//	config.Validate(merr)
//	assert.NoError(t, merr.Err())
//}
//
//// TestHealthSettings_Validate tests validating a HealthSettings. Validation should always pass.
//func TestHealthSettings_Validate(t *testing.T) {
//	merr := errors.NewMultiError("test")
//	config := HealthSettings{}
//	config.Validate(merr)
//	assert.NoError(t, merr.Err())
//}
//
//// Test validating the ListenSettings successfully.
//func TestListenSettings_Validate_Ok(t *testing.T) {
//	var testTable = []struct {
//		desc   string
//		config ListenSettings
//	}{
//		{
//			desc: "listen enabled, small buffer",
//			config: ListenSettings{
//				Enabled: true,
//				Buffer:  1,
//			},
//		},
//		{
//			desc: "listen enabled, larger buffer",
//			config: ListenSettings{
//				Enabled: true,
//				Buffer:  100,
//			},
//		},
//		{
//			desc: "listen disabled, small buffer",
//			config: ListenSettings{
//				Enabled: false,
//				Buffer:  1,
//			},
//		},
//		{
//			desc: "listen disabled, larger buffer",
//			config: ListenSettings{
//				Enabled: false,
//				Buffer:  100,
//			},
//		},
//	}
//
//	for _, testCase := range testTable {
//		merr := errors.NewMultiError("test")
//
//		testCase.config.Validate(merr)
//		assert.NoError(t, merr.Err(), testCase.desc)
//	}
//}
//
//// Test validating the ListenSettings unsuccessfully.
//func TestListenSettings_Validate_Error(t *testing.T) {
//	var testTable = []struct {
//		desc     string
//		errCount int
//		config   ListenSettings
//	}{
//		{
//			desc:     "listen enabled, zero buffer",
//			errCount: 1,
//			config: ListenSettings{
//				Enabled: true,
//				Buffer:  0,
//			},
//		},
//		{
//			desc:     "listen enabled, negative buffer",
//			errCount: 1,
//			config: ListenSettings{
//				Enabled: true,
//				Buffer:  -1,
//			},
//		},
//		{
//			desc:     "listen disabled, zero buffer",
//			errCount: 1,
//			config: ListenSettings{
//				Enabled: false,
//				Buffer:  0,
//			},
//		},
//		{
//			desc:     "listen disabled, negative buffer",
//			errCount: 1,
//			config: ListenSettings{
//				Enabled: false,
//				Buffer:  -1,
//			},
//		},
//	}
//
//	for _, testCase := range testTable {
//		merr := errors.NewMultiError("test")
//
//		testCase.config.Validate(merr)
//		assert.Error(t, merr.Err(), testCase.desc)
//		assert.Equal(t, testCase.errCount, len(merr.Errors), merr.Error())
//	}
//}
