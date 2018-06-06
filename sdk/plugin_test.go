package sdk

// File level global test configuration.
//var testConfig = config.PluginConfig{
//	SchemeVersion: config.SchemeVersion{
//		Version: "1.0",
//	},
//	Network: &config.NetworkSettings{
//		Type:    "tcp",
//		Address: "test_config",
//	},
//	Settings: &config.PluginSettings{
//		Read:        &config.ReadSettings{Buffer: 1024},
//		Write:       &config.WriteSettings{Buffer: 1024},
//		Transaction: &config.TransactionSettings{TTL: "2s"},
//	},
//}

//// TestNewPluginNilHandlers tests creating a new Plugin with nil handlers
//func TestNewPluginNilHandlers(t *testing.T) {
//	_, err := NewPlugin(nil, &testConfig)
//	assert.Error(t, err)
//}
//
//// TestNewPlugin tests creating a new Plugin with the required handlers defined.
//func TestNewPlugin(t *testing.T) {
//	// Create valid handlers for the Plugin.
//	h, err := NewHandlers(testDeviceIdentifier, nil)
//	assert.NoError(t, err)
//
//	// Create the plugin.
//	p, err := NewPlugin(h, &testConfig)
//	assert.NoError(t, err)
//
//	assert.Nil(t, p.server, "server should not be initialized with new plugin")
//	assert.Nil(t, p.dataManager, "data manager should not be initialized with new plugin")
//	assert.Nil(t, p.handlers.DeviceEnumerator)
//
//	assert.Equal(t, &h.DeviceIdentifier, &p.handlers.DeviceIdentifier)
//	assert.NotNil(t, p.Config, "plugin should be configured on init")
//}
//
//// TestNewPluginWithConfig tests creating a new Plugin and passing a valid
//// config in as an argument.
//func TestNewPluginWithConfig(t *testing.T) {
//	// Create valid handlers for the Plugin.
//	h, err := NewHandlers(testDeviceIdentifier, nil)
//	assert.NoError(t, err)
//
//	// Create a configuration for the Plugin.
//	c := config.PluginConfig{
//		Version: "1.0",
//		Network: config.NetworkSettings{
//			Type:    "tcp",
//			Address: ":666",
//		},
//	}
//
//	// Create the plugin.
//	p, err := NewPlugin(h, &c)
//	assert.NoError(t, err)
//	assert.NotNil(t, p.Config, "plugin should be configured")
//}
//
//// TestNewPluginWithIncompleteConfig tests creating a new Plugin and passing in
//// an incomplete PluginConfig instance as an argument. The constructor should not
//// return an error with a bad/incomplete config.
//func TestNewPluginWithIncompleteConfig(t *testing.T) {
//	// Create valid handlers for the Plugin.
//	h, err := NewHandlers(testDeviceIdentifier, nil)
//	assert.NoError(t, err)
//
//	// network spec missing but required
//	c := config.PluginConfig{
//		Version: "1.0",
//	}
//
//	// Create the plugin.
//	_, err = NewPlugin(h, &c)
//	assert.NoError(t, err)
//}
//
//// TestNewPluginNilConfig tests creating a new Plugin, passing in nil as the
//// configuration. This should cause the plugin to search for plugin configuration
//// files to load config from. Here we expect it to error out because it should
//// not find a config file.
//func TestNewPluginNilConfig(t *testing.T) {
//	h, err := NewHandlers(testDeviceIdentifier, nil)
//	assert.NoError(t, err)
//
//	// Create the plugin
//	plugin, err := NewPlugin(h, nil)
//	assert.Nil(t, plugin)
//	assert.Error(t, err)
//}
//
//// TestPlugin_RegisterDeviceHandlers tests registering device handlers when
//// none are already registered.
//func TestPlugin_RegisterDeviceHandlers(t *testing.T) {
//	devHandler1 := DeviceHandler{Type: "test", Model: "model1"}
//	devHandler2 := DeviceHandler{Type: "test", Model: "model2"}
//
//	plugin := Plugin{}
//	assert.Nil(t, plugin.deviceHandlers)
//
//	plugin.RegisterDeviceHandlers(&devHandler1, &devHandler2)
//
//	assert.NotNil(t, plugin.deviceHandlers)
//	assert.Equal(t, 2, len(plugin.deviceHandlers))
//}
//
//// TestPlugin_RegisterDeviceHandlers tests registering device handlers when
//// some are already registered.
//func TestPlugin_RegisterDeviceHandlers2(t *testing.T) {
//	devHandler1 := DeviceHandler{Type: "test", Model: "model1"}
//	devHandler2 := DeviceHandler{Type: "test", Model: "model2"}
//
//	plugin := Plugin{}
//	plugin.deviceHandlers = []*DeviceHandler{&devHandler1}
//	assert.Equal(t, 1, len(plugin.deviceHandlers))
//
//	plugin.RegisterDeviceHandlers(&devHandler2)
//
//	assert.NotNil(t, plugin.deviceHandlers)
//	assert.Equal(t, 2, len(plugin.deviceHandlers))
//}
//
//// TestPlugin_RegisterDeviceSetupActions tests registering device setup actions for a filter
//// when the device setup actions map doesn't exist.
//func TestPlugin_RegisterDeviceSetupActions(t *testing.T) {
//	plugin := Plugin{}
//	setupFn := func(p *Plugin, d *Device) error { return nil }
//
//	assert.Nil(t, plugin.deviceSetupActions)
//
//	plugin.RegisterDeviceSetupActions("type=test", setupFn)
//
//	assert.NotNil(t, plugin.deviceSetupActions)
//	assert.Equal(t, 1, len(plugin.deviceSetupActions))
//	assert.Equal(t, 1, len(plugin.deviceSetupActions["type=test"]))
//}
//
//// TestPlugin_RegisterDeviceSetupActions2 tests registering device setup actions for a filter
//// when the device setup actions map does exist, but the filter does not already exist.
//func TestPlugin_RegisterDeviceSetupActions2(t *testing.T) {
//	plugin := Plugin{}
//	plugin.deviceSetupActions = make(map[string][]deviceAction)
//	setupFn := func(p *Plugin, d *Device) error { return nil }
//
//	assert.NotNil(t, plugin.deviceSetupActions)
//
//	plugin.RegisterDeviceSetupActions("type=test", setupFn)
//
//	assert.NotNil(t, plugin.deviceSetupActions)
//	assert.Equal(t, 1, len(plugin.deviceSetupActions))
//	assert.Equal(t, 1, len(plugin.deviceSetupActions["type=test"]))
//}
//
//// TestPlugin_RegisterDeviceSetupActions3 tests registering device setup actions for a filter
//// when the device setup actions map exists and the filter already exists in the map.
//func TestPlugin_RegisterDeviceSetupActions3(t *testing.T) {
//	plugin := Plugin{}
//	setupFn1 := func(p *Plugin, d *Device) error { return nil }
//	setupFn2 := func(p *Plugin, d *Device) error { return nil }
//	plugin.deviceSetupActions = make(map[string][]deviceAction)
//	plugin.deviceSetupActions["type=test"] = []deviceAction{setupFn1}
//
//	assert.NotNil(t, plugin.deviceSetupActions)
//	assert.Equal(t, 1, len(plugin.deviceSetupActions))
//	assert.Equal(t, 1, len(plugin.deviceSetupActions["type=test"]))
//
//	plugin.RegisterDeviceSetupActions("type=test", setupFn2)
//
//	assert.NotNil(t, plugin.deviceSetupActions)
//	assert.Equal(t, 1, len(plugin.deviceSetupActions))
//	assert.Equal(t, 2, len(plugin.deviceSetupActions["type=test"]))
//}
//
//// TestPlugin_RegisterPostRunActions tests registering post run actions
//// when none are already defined.
//func TestPlugin_RegisterPostRunActions(t *testing.T) {
//	action := func(p *Plugin) error { return nil }
//	plugin := Plugin{}
//	assert.Nil(t, plugin.postRunActions)
//
//	plugin.RegisterPostRunActions(action)
//
//	assert.NotNil(t, plugin.postRunActions)
//	assert.Equal(t, 1, len(plugin.postRunActions))
//}
//
//// TestPlugin_RegisterPostRunActions2 tests registering post run actions
//// when some are already defined.
//func TestPlugin_RegisterPostRunActions2(t *testing.T) {
//	action1 := func(p *Plugin) error { return nil }
//	action2 := func(p *Plugin) error { return nil }
//	plugin := Plugin{}
//
//	plugin.postRunActions = []pluginAction{action1}
//	assert.NotNil(t, plugin.postRunActions)
//	assert.Equal(t, 1, len(plugin.postRunActions))
//
//	plugin.RegisterPostRunActions(action2)
//
//	assert.NotNil(t, plugin.postRunActions)
//	assert.Equal(t, 2, len(plugin.postRunActions))
//}
//
//// TestPlugin_RegisterPreRunActions tests registering pre run actions
//// when none are already defined.
//func TestPlugin_RegisterPreRunActions(t *testing.T) {
//	action := func(p *Plugin) error { return nil }
//	plugin := Plugin{}
//	assert.Nil(t, plugin.preRunActions)
//
//	plugin.RegisterPreRunActions(action)
//
//	assert.NotNil(t, plugin.preRunActions)
//	assert.Equal(t, 1, len(plugin.preRunActions))
//}
//
//// TestPlugin_RegisterPreRunActions2 tests registering pre run actions
//// when some are already defined.
//func TestPlugin_RegisterPreRunActions2(t *testing.T) {
//	action1 := func(p *Plugin) error { return nil }
//	action2 := func(p *Plugin) error { return nil }
//	plugin := Plugin{}
//
//	plugin.preRunActions = []pluginAction{action1}
//	assert.NotNil(t, plugin.preRunActions)
//	assert.Equal(t, 1, len(plugin.preRunActions))
//
//	plugin.RegisterPreRunActions(action2)
//
//	assert.NotNil(t, plugin.preRunActions)
//	assert.Equal(t, 2, len(plugin.preRunActions))
//
//}
//
//// TestPlugin_logInfo tests logging out the plugin info.
//func TestPlugin_logInfo(t *testing.T) {
//	h, err := NewHandlers(testDeviceIdentifier, nil)
//	assert.NoError(t, err)
//
//	// Create the plugin.
//	p, err := NewPlugin(h, &testConfig)
//	assert.NoError(t, err)
//	assert.NotNil(t, p)
//
//	// Should not cause any kind of error
//	p.logInfo()
//}
//
//// TestPlugin_Configure tests configuring a Plugin using a config file
//// specified via environment variable.
//func TestPlugin_Configure(t *testing.T) {
//	cfgFilePath := "tmp/config.yml"
//	cfg := `
//version: 1.0.0
//debug: true
//network:
//  type: tcp
//  address: ":50051"
//settings:
//  loop_delay: 100
//  read:
//    buffer_size: 150
//  write:
//    buffer_size: 150
//    per_loop: 4
//  transaction:
//    ttl: 600s`
//
//	err := writeConfigFile(cfgFilePath, cfg)
//	assert.NoError(t, err)
//
//	defer func() {
//		err = os.RemoveAll("tmp")
//		assert.NoError(t, err)
//	}()
//
//	test.CheckErr(t, os.Setenv("PLUGIN_CONFIG", "tmp"))
//
//	// Create valid handlers for the Plugin.
//	h, err := NewHandlers(testDeviceIdentifier, nil)
//	assert.NoError(t, err)
//
//	// Create the plugin.
//	p, err := NewPlugin(h, nil)
//	assert.NoError(t, err)
//
//	assert.NotNil(t, p.Config, "plugin is not configured, but should be")
//	assert.Equal(t, "1.0.0", p.Config.Version)
//}
//
//// TestPlugin_setup tests setting up a Plugin successfully. This means that
//// the state is validated, the devices are registered, and the Plugin components
//// (server, data manager) are created.
//func TestPlugin_setup(t *testing.T) {
//	// Create valid handlers for the Plugin.
//	h, err := NewHandlers(testDeviceIdentifier, testDeviceEnumerator)
//	assert.NoError(t, err)
//
//	// Create the plugin.
//	p, err := NewPlugin(h, &testConfig)
//	assert.NoError(t, err)
//
//	// CONSIDER: Can we move setup functionality to the constructor?
//	err = p.setup()
//	assert.NoError(t, err)
//
//	assert.NotNil(t, p.server, "server should be initialized on setup")
//	assert.NotNil(t, p.dataManager, "data manager should be initialized on setup")
//}
//
//// TestPlugin_setup2 tests setting up a Plugin unsuccessfully. This means that
//// the state is validated, the devices are registered, and the Plugin components
//// (server, data manager) are created. In this case, handler validation should fail.
//func TestPlugin_setup2(t *testing.T) {
//	// Create invalid handlers for the plugin.
//	h := Handlers{}
//	p, err := NewPlugin(&h, &testConfig)
//	assert.NoError(t, err)
//
//	err = p.setup()
//	assert.Error(t, err)
//}
//
//// TestPlugin_setup3 tests setting up a plugin unsuccessfully. It should fail
//// setup validation due to a nil config.
//func TestPlugin_setup3(t *testing.T) {
//	h, err := NewHandlers(testDeviceIdentifier, testDeviceEnumerator)
//	assert.NoError(t, err)
//
//	p := Plugin{
//		handlers: h,
//	}
//
//	err = p.setup()
//	assert.Error(t, err)
//}
//
//// TestPlugin_setup4 tests setting up a plugin unsuccessfully. It should fail
//// setup validation due to a bad configuration.
//func TestPlugin_setup4(t *testing.T) {
//	h, err := NewHandlers(testDeviceIdentifier, testDeviceEnumerator)
//	assert.NoError(t, err)
//
//	p, err := NewPlugin(h, &testConfig)
//	assert.NoError(t, err)
//
//	// make the plugin config bad
//	p.Config.Settings.Transaction.TTL = "foo"
//
//	err = p.setup()
//	assert.Error(t, err)
//}
