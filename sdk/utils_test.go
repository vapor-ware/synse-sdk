package sdk

//
//import (
//	"fmt"
//	"sort"
//	"testing"
//	"time"
//
//	"github.com/stretchr/testify/assert"
//)
//
//// TestMakeIDString tests making a compound ID string out of the device
//// identifier components (rack, board, device).
//func TestMakeIDString(t *testing.T) {
//	var makeIDStringTestTable = []struct {
//		rack     string
//		board    string
//		device   string
//		expected string
//	}{
//		{
//			rack:     "rack",
//			board:    "board",
//			device:   "device",
//			expected: "rack-board-device",
//		},
//		{
//			rack:     "123",
//			board:    "456",
//			device:   "789",
//			expected: "123-456-789",
//		},
//		{
//			rack:     "abc",
//			board:    "def",
//			device:   "ghi",
//			expected: "abc-def-ghi",
//		},
//		{
//			rack:     "1234567890abcdefghi",
//			board:    "1",
//			device:   "2",
//			expected: "1234567890abcdefghi-1-2",
//		},
//		{
//			rack:     "-----",
//			board:    "_____",
//			device:   "+=+=&8^",
//			expected: "------_____-+=+=&8^",
//		},
//	}
//
//	for _, tc := range makeIDStringTestTable {
//		actual := makeIDString(tc.rack, tc.board, tc.device)
//		assert.Equal(t, tc.expected, actual)
//	}
//}
//
//// TestNewUID tests creating new device UIDs successfully.
//func TestNewUID(t *testing.T) {
//	var newUIDTestTable = []struct {
//		p        string
//		d        string
//		m        string
//		c        string
//		expected string
//	}{
//		{
//			p:        "test-protocol",
//			d:        "test-device",
//			m:        "test-model",
//			c:        "test-comp",
//			expected: "732bb43a825b8330e6d50a6722a8e1f0",
//		},
//		{
//			p:        "i2c",
//			d:        "thermistor",
//			m:        "max116",
//			c:        "1",
//			expected: "019de8ff9de6aba9ddb9ebb6d5f5b5e0",
//		},
//		{
//			p:        "",
//			d:        "",
//			m:        "",
//			c:        "",
//			expected: "d41d8cd98f00b204e9800998ecf8427e",
//		},
//		{
//			p:        "?",
//			d:        "!",
//			m:        "%",
//			c:        "$",
//			expected: "65722f8565fb36c7a6da67bae4ee1f2d",
//		},
//	}
//
//	for _, tc := range newUIDTestTable {
//		actual := newUID(tc.p, tc.d, tc.m, tc.c)
//		assert.Equal(t, tc.expected, actual)
//	}
//}
//
//// TestFilterDevices tests filtering devices successfully. These cases
//// exercise valid filter strings against different combinations of filters
//// and values.
//func TestFilterDevices(t *testing.T) {
//	dev1 := &Device{
//		Kind:    "foo.temperature",
//		Handler: &DeviceHandler{},
//	}
//	dev2 := &Device{
//		Kind:    "bar.temperature",
//		Handler: &DeviceHandler{},
//	}
//	dev3 := &Device{
//		Kind:    "baz.pressure",
//		Handler: &DeviceHandler{},
//	}
//
//	defer resetContext()
//
//	// Populate the device map with the test devices.
//	ctx.devices = map[string]*Device{
//		"dev1": dev1,
//		"dev2": dev2,
//		"dev3": dev3,
//	}
//
//	// Set up the test cases
//	var filterDevicesTestTable = []struct {
//		desc     string
//		filter   string
//		expected []*Device
//	}{
//		{
//			desc:     "devices with type temperature",
//			filter:   "type=temperature",
//			expected: []*Device{dev1, dev2},
//		},
//		{
//			desc:     "devices with type pressure",
//			filter:   "type=pressure",
//			expected: []*Device{dev3},
//		},
//		{
//			desc:     "devices with kind baz.pressure",
//			filter:   "kind=baz.pressure",
//			expected: []*Device{dev3},
//		},
//		{
//			desc:     "devices with type pressure and type temperature (can't have two types)",
//			filter:   "type=pressure,type=temperature",
//			expected: []*Device{},
//		},
//		{
//			desc:     "devices with type none (should not match any)",
//			filter:   "type=none",
//			expected: []*Device{},
//		},
//		{
//			desc:     "devices with type temperature and kind foo.temperature",
//			filter:   "type=temperature,kind=foo.temperature",
//			expected: []*Device{dev1},
//		},
//		{
//			desc:     "devices with type pressure and kind foo.temperature",
//			filter:   "type=pressure,kind=foo.temperature",
//			expected: []*Device{},
//		},
//		{
//			desc:     "devices of any type",
//			filter:   "type=*",
//			expected: []*Device{dev1, dev2, dev3},
//		},
//		{
//			desc:     "devices of any kind",
//			filter:   "kind=*",
//			expected: []*Device{dev1, dev2, dev3},
//		},
//		{
//			desc:     "devices of any kind with type temperature",
//			filter:   "type=temperature,kind=*",
//			expected: []*Device{dev1, dev2},
//		},
//		{
//			desc:     "devices of any type with kind baz.pressure",
//			filter:   "type=*,kind=baz.pressure",
//			expected: []*Device{dev3},
//		},
//	}
//
//	for _, testCase := range filterDevicesTestTable {
//		actual, err := filterDevices(testCase.filter)
//		expected := testCase.expected
//		assert.NoError(t, err, testCase.desc)
//		assert.Equal(t, len(expected), len(actual), testCase.desc)
//
//		// Sort the expected and actual (we sort by Kind for the tests
//		// since it is unique for each test device)
//		sort.SliceStable(expected, func(i int, j int) bool { return expected[i].Kind < expected[j].Kind })
//		sort.SliceStable(actual, func(i int, j int) bool { return actual[i].Kind < actual[j].Kind })
//		assert.Equal(t, testCase.expected, actual, "filter: %v", testCase.filter, testCase.desc)
//	}
//}
//
//// TestFilterDevicesErr tests filtering devices when the given filter
//// string results in a filtering error.
//func TestFilterDevicesErr(t *testing.T) {
//	dev1 := &Device{
//		Kind:    "foo.temperature",
//		Handler: &DeviceHandler{},
//	}
//	dev2 := &Device{
//		Kind:    "bar.temperature",
//		Handler: &DeviceHandler{},
//	}
//
//	defer resetContext()
//
//	// Populate the device map with the test devices.
//	ctx.devices = map[string]*Device{
//		"dev1": dev1,
//		"dev2": dev2,
//	}
//
//	// Set up the test cases
//	var filterDevicesTestTable = []string{
//		// no filter - when filtering, we should always have a filter
//		// string specified when calling filterDevices
//		"",
//
//		// unsupported filter keys
//		"invalid=temperature",
//		"KIND=bar",
//		"Type=temperature",
//	}
//
//	for _, testCase := range filterDevicesTestTable {
//		_, err := filterDevices(testCase)
//		assert.Error(t, err)
//	}
//}
//
//// TestParseRFC3339_Ok tests successfully parsing an RFC3339 timestamp
//// into a Time struct.
//func TestParseRFC3339_Ok(t *testing.T) {
//	// get the EST location
//	est, err := time.LoadLocation("EST")
//	assert.NoError(t, err)
//
//	var tests = []struct {
//		timestamp string
//		expected  time.Time
//	}{
//		{
//			// no timestamp defaults to zero value for time
//			timestamp: "",
//			expected:  time.Time{},
//		},
//		{
//			// rfc3339 utc
//			timestamp: "2018-10-16T18:22:50Z",
//			expected:  time.Date(2018, 10, 16, 18, 22, 50, 0, time.UTC),
//		},
//		{
//			// rfc3339nano utc
//			timestamp: "2018-10-16T18:22:50.573971054Z",
//			expected:  time.Date(2018, 10, 16, 18, 22, 50, 573971054, time.UTC),
//		},
//		{
//			// rcf3339 est
//			timestamp: "2018-10-16T13:25:00-05:00",
//			expected:  time.Date(2018, 10, 16, 13, 25, 0, 0, est),
//		},
//		{
//			// rfc3339nano est
//			timestamp: "2018-10-16T13:25:00.410241272-05:00",
//			expected:  time.Date(2018, 10, 16, 13, 25, 0, 410241272, est),
//		},
//	}
//
//	for i, tt := range tests {
//		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
//			actual, err := ParseRFC3339(tt.timestamp)
//			assert.NoError(t, err)
//			assert.True(t, tt.expected.Equal(actual))
//		})
//	}
//}
//
//// TestParseRFC3339_Error tests unsuccessfully parsing a timestamp into
//// a Time struct
//func TestParseRFC3339_Error(t *testing.T) {
//	var tests = []string{
//		"foobar",
//		"...",
//		"16 Oct 18 18:22 UTC",
//		"16 Oct 18 13:25 EST",
//		"Tue Oct 16 18:22:50 2018",
//		"Tue Oct 16 13:25:00 2018",
//		"6:22PM",
//		"1:25PM",
//		"Tue, 16 Oct 2018 18:22:50 +0000",
//		"Tue, 16 Oct 2018 13:25:00 -0500",
//		"Oct 16 18:22:50.573997",
//		"Oct 16 13:25:00.410271",
//		"Tue Oct 16 18:22:50 UTC 2018",
//		"Tue Oct 16 13:25:00 EST 2018",
//	}
//
//	for i, tt := range tests {
//		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
//			actual, err := ParseRFC3339(tt)
//			assert.Error(t, err)
//			assert.Empty(t, actual)
//		})
//	}
//}
//
//// TestGetCurrentTime tests getting the current time.
//func TestGetCurrentTime(t *testing.T) {
//	// TODO: figure out how to test the response...
//	out := GetCurrentTime()
//	assert.NotEmpty(t, out)
//}
//
//// Test_getTypeByNameOk tests getting a type that exists.
//func Test_getTypeByNameOk(t *testing.T) {
//	defer resetContext()
//
//	ctx.outputTypes["foo"] = &OutputType{Name: "foo"}
//
//	ot, err := GetTypeByName("foo")
//	assert.NoError(t, err)
//	assert.NotNil(t, ot)
//	assert.Equal(t, "foo", ot.Name)
//}
//
//// Test_getTypeByNameErr tests getting a type that doesn't exist.
//func Test_getTypeByNameErr(t *testing.T) {
//	ot, err := GetTypeByName("bar")
//	assert.Error(t, err)
//	assert.Nil(t, ot)
//}
//
//// Test_logStartupInfo tests logging out info which is done on startup.
//// There isn't much to check here other than it runs and completes without
//// any issues.
//func Test_logStartupInfo(t *testing.T) {
//	logStartupInfo()
//}
//
//// Test_registerDevices tests registering devices with the plugin when there
//// are no devices to register.
//func Test_registerDevices(t *testing.T) {
//	defer func() {
//		resetContext()
//		Config.reset()
//	}()
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		DynamicRegistration: &DynamicRegistrationSettings{
//			Config: []map[string]interface{}{
//				{},
//			},
//		},
//	}
//	Config.Device = &DeviceConfig{
//		Devices: []*DeviceKind{},
//	}
//
//	assert.Equal(t, 0, len(ctx.devices))
//	err := registerDevices()
//	assert.NoError(t, err)
//	assert.Equal(t, 0, len(ctx.devices))
//}
//
//// Test_registerDevices2 tests registering devices with the plugin when
//// dynamic registration results in an error.
//func Test_registerDevices2(t *testing.T) {
//	defer func() {
//		resetContext()
//		Config.reset()
//	}()
//
//	ctx.dynamicDeviceRegistrar = func(i map[string]interface{}) ([]*Device, error) {
//		return nil, fmt.Errorf("test")
//	}
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		DynamicRegistration: &DynamicRegistrationSettings{
//			Config: []map[string]interface{}{
//				{},
//			},
//		},
//	}
//	Config.Device = &DeviceConfig{
//		Devices: []*DeviceKind{},
//	}
//
//	assert.Equal(t, 0, len(ctx.devices))
//	err := registerDevices()
//	assert.Error(t, err)
//	assert.Equal(t, 0, len(ctx.devices))
//}
//
//// Test_registerDevices3 tests registering devices with the plugin when there
//// is a device config, but it is invalid.
//func Test_registerDevices3(t *testing.T) {
//	defer func() {
//		resetContext()
//		Config.reset()
//	}()
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		DynamicRegistration: &DynamicRegistrationSettings{
//			Config: []map[string]interface{}{
//				{},
//			},
//		},
//	}
//	Config.Device = &DeviceConfig{
//		Devices: []*DeviceKind{
//			{
//				Name: "foo",
//				Instances: []*DeviceInstance{
//					{
//						Location: "bar",
//					},
//				},
//			},
//		},
//	}
//
//	assert.Equal(t, 0, len(ctx.devices))
//	err := registerDevices()
//	assert.Error(t, err)
//	assert.Equal(t, 0, len(ctx.devices))
//}
//
//// Test_registerDevices4 tests registering devices with the plugin when there
//// is a device config and it is valid.
//func Test_registerDevices4(t *testing.T) {
//	defer func() {
//		resetContext()
//		Config.reset()
//	}()
//
//	ctx.deviceHandlers = []*DeviceHandler{
//		{Name: "foo"},
//	}
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		DynamicRegistration: &DynamicRegistrationSettings{
//			Config: []map[string]interface{}{
//				{},
//			},
//		},
//	}
//	Config.Device = &DeviceConfig{
//		Locations: []*LocationConfig{
//			{
//				Name:  "bar",
//				Rack:  &LocationData{Name: "rack"},
//				Board: &LocationData{Name: "board"},
//			},
//		},
//		Devices: []*DeviceKind{
//			{
//				Name: "foo",
//				Instances: []*DeviceInstance{
//					{
//						Location: "bar",
//					},
//				},
//			},
//		},
//	}
//
//	assert.Equal(t, 0, len(ctx.devices))
//	err := registerDevices()
//	assert.NoError(t, err)
//	assert.Equal(t, 1, len(ctx.devices))
//}
