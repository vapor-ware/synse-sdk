package sdk

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vapor-ware/synse-sdk/internal/test"
	"github.com/vapor-ware/synse-sdk/sdk/config"
)

// TestMakeDevices tests making devices where two instances match one prototype.
func TestMakeDevices(t *testing.T) {
	inst := []*config.DeviceConfig{&testDeviceConfig1, &testDeviceConfig2}
	proto := []*config.PrototypeConfig{&testPrototypeConfig1}

	devices, err := makeDevices(inst, proto, &testPlugin)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(devices))
}

// TestMakeDevices2 tests making devices when no instances match the prototype.
func TestMakeDevices2(t *testing.T) {
	inst := []*config.DeviceConfig{&testDeviceConfig1, &testDeviceConfig2}
	proto := []*config.PrototypeConfig{&testPrototypeConfig2}

	devices, err := makeDevices(inst, proto, &testPlugin)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(devices))
}

// TestMakeDevices3 tests making devices when one instance matches one of two prototypes.
func TestMakeDevices3(t *testing.T) {
	inst := []*config.DeviceConfig{&testDeviceConfig1}
	proto := []*config.PrototypeConfig{&testPrototypeConfig1, &testPrototypeConfig2}

	devices, err := makeDevices(inst, proto, &testPlugin)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(devices))
}

// TestMakeDevices4 tests making devices when no prototypes exist for instances to
// match with.
func TestMakeDevices4(t *testing.T) {
	inst := []*config.DeviceConfig{&testDeviceConfig1, &testDeviceConfig2}
	var proto []*config.PrototypeConfig

	devices, err := makeDevices(inst, proto, &testPlugin)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(devices))
}

// TestMakeDevices5 tests making devices when no instances exist for protocols to
// be matched with.
func TestMakeDevices5(t *testing.T) {
	var inst []*config.DeviceConfig
	proto := []*config.PrototypeConfig{&testPrototypeConfig1, &testPrototypeConfig2}

	devices, err := makeDevices(inst, proto, &testPlugin)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(devices))
}

// TestMakeDevices6 tests making devices when the plugin is incorrectly configured
// (no device identifier handler), which should prohibit devices from being created.
func TestMakeDevices6(t *testing.T) {
	inst := []*config.DeviceConfig{&testDeviceConfig1, &testDeviceConfig2}
	proto := []*config.PrototypeConfig{&testPrototypeConfig1}

	plugin := makeTestPlugin()
	plugin.handlers.DeviceIdentifier = nil

	_, err := makeDevices(inst, proto, plugin)
	assert.Error(t, err)
}

// TestMakeDevices7 tests making devices when the plugin is incorrectly configured
// (no device handlers), which should prohibit devices from being created.
func TestMakeDevices7(t *testing.T) {
	inst := []*config.DeviceConfig{&testDeviceConfig1, &testDeviceConfig2}
	proto := []*config.PrototypeConfig{&testPrototypeConfig1}

	plugin := makeTestPlugin()
	plugin.deviceHandlers = []*DeviceHandler{}

	_, err := makeDevices(inst, proto, plugin)
	assert.Error(t, err)
}

// TestSetupSocket tests setting up the socket when the socket path
// already exists.
func TestSetupSocket(t *testing.T) {
	// Set up a temporary directory for testing
	dir, err := ioutil.TempDir("", "testing")
	assert.NoError(t, err)
	defer func() {
		test.CheckErr(t, os.RemoveAll(dir))
	}()

	// Set the socket path to the temp dir for the test
	sockPath = dir

	sock, err := setupSocket("test.sock")
	if err != nil {
		t.Error(err)
	}

	// Verify the socket path+name
	assert.Equal(t, filepath.Join(dir, "test.sock"), sock)

	// Verify that the socket path exists
	_, err = os.Stat(sockPath)
	assert.NoError(t, err)
}

// TestSetupSocket2 tests setting up the socket when the socket path
// does not already exists.
func TestSetupSocket2(t *testing.T) {
	// Set up a temporary directory for testing
	dir, err := ioutil.TempDir("", "testing")
	assert.NoError(t, err)
	// remove the temp dir now - it shouldn't exist when we set up the socket
	test.CheckErr(t, os.RemoveAll(dir))

	// Set the socket path to the temp dir for the test
	sockPath = dir

	sock, err := setupSocket("test.sock")
	if err != nil {
		t.Error(err)
	}

	// Verify the socket path+name
	assert.Equal(t, filepath.Join(dir, "test.sock"), sock)

	// Verify that the socket path exists
	_, err = os.Stat(sockPath)
	assert.NoError(t, err)
}

// TestSetupSocket2 tests setting up the socket when the socket path
// and the socket itself already exist.
func TestSetupSocket3(t *testing.T) {
	// Set up a temporary directory for testing
	dir, err := ioutil.TempDir("", "testing")
	assert.NoError(t, err)
	defer func() {
		test.CheckErr(t, os.RemoveAll(dir))
	}()

	// Set the socket path to the temp dir for the test
	sockPath = dir

	// Make the socket file
	filename := filepath.Join(dir, "test.sock")
	_, err = os.Create(filename)
	assert.NoError(t, err)

	sock, err := setupSocket("test.sock")
	if err != nil {
		t.Error(err)
	}

	// Verify the socket path+name
	assert.Equal(t, filepath.Join(dir, "test.sock"), sock)

	// Verify that the socket path exists
	_, err = os.Stat(sockPath)
	assert.NoError(t, err)

	// Verify that the socket itself no longer exists (setupSocket cleans
	// up old socket instances)
	_, err = os.Stat(filename)
	exists := !os.IsNotExist(err)
	assert.False(t, exists, "socket should no longer exist, but still does")
}

// TestMakeIDString tests making a compound ID string out of the device
// identifier components (rack, board, device).
func TestMakeIDString(t *testing.T) {
	var makeIDStringTestTable = []struct {
		rack     string
		board    string
		device   string
		expected string
	}{
		{
			rack:     "rack",
			board:    "board",
			device:   "device",
			expected: "rack-board-device",
		},
		{
			rack:     "123",
			board:    "456",
			device:   "789",
			expected: "123-456-789",
		},
		{
			rack:     "abc",
			board:    "def",
			device:   "ghi",
			expected: "abc-def-ghi",
		},
		{
			rack:     "1234567890abcdefghi",
			board:    "1",
			device:   "2",
			expected: "1234567890abcdefghi-1-2",
		},
		{
			rack:     "-----",
			board:    "_____",
			device:   "+=+=&8^",
			expected: "------_____-+=+=&8^",
		},
	}

	for _, tc := range makeIDStringTestTable {
		actual := makeIDString(tc.rack, tc.board, tc.device)
		assert.Equal(t, tc.expected, actual)
	}
}

// TestNewUID tests creating new device UIDs successfully.
func TestNewUID(t *testing.T) {
	var newUIDTestTable = []struct {
		p        string
		d        string
		m        string
		c        string
		expected string
	}{
		{
			p:        "test-protocol",
			d:        "test-device",
			m:        "test-model",
			c:        "test-comp",
			expected: "732bb43a825b8330e6d50a6722a8e1f0",
		},
		{
			p:        "i2c",
			d:        "thermistor",
			m:        "max116",
			c:        "1",
			expected: "019de8ff9de6aba9ddb9ebb6d5f5b5e0",
		},
		{
			p:        "",
			d:        "",
			m:        "",
			c:        "",
			expected: "d41d8cd98f00b204e9800998ecf8427e",
		},
		{
			p:        "?",
			d:        "!",
			m:        "%",
			c:        "$",
			expected: "65722f8565fb36c7a6da67bae4ee1f2d",
		},
	}

	for _, tc := range newUIDTestTable {
		actual := newUID(tc.p, tc.d, tc.m, tc.c)
		assert.Equal(t, tc.expected, actual)
	}
}

// TestFilterDevices tests filtering devices successfully. These cases
// exercise valid filter strings against different combinations of filters
// and values.
func TestFilterDevices(t *testing.T) {
	dev1 := &Device{
		Type:    "temperature",
		Model:   "foo",
		Handler: &DeviceHandler{},
	}
	dev2 := &Device{
		Type:    "temperature",
		Model:   "bar",
		Handler: &DeviceHandler{},
	}
	dev3 := &Device{
		Type:    "pressure",
		Model:   "baz",
		Handler: &DeviceHandler{},
	}

	// Populate the device map with the test devices.
	deviceMap = map[string]*Device{
		"dev1": dev1,
		"dev2": dev2,
		"dev3": dev3,
	}

	// Set up the test cases
	var filterDevicesTestTable = []struct {
		filter   string
		expected []*Device
	}{
		{
			// devices with type temperature
			filter:   "type=temperature",
			expected: []*Device{dev1, dev2},
		},
		{
			// devices with type pressure
			filter:   "type=pressure",
			expected: []*Device{dev3},
		},
		{
			// devices with model baz
			filter:   "model=baz",
			expected: []*Device{dev3},
		},
		{
			// devices with type pressure and type temperature (can't have two types)
			filter:   "type=pressure,type=temperature",
			expected: []*Device{},
		},
		{
			// devices with type none (should not match any)
			filter:   "type=none",
			expected: []*Device{},
		},
		{
			// devices with type temperature and model foo
			filter:   "type=temperature,model=foo",
			expected: []*Device{dev1},
		},
		{
			// devices with type pressure and model foo
			filter:   "type=pressure,model=foo",
			expected: []*Device{},
		},
		{
			// devices of any type
			filter:   "type=*",
			expected: []*Device{dev1, dev2, dev3},
		},
		{
			// devices of any model
			filter:   "model=*",
			expected: []*Device{dev1, dev2, dev3},
		},
		{
			// devices of any model with type temperature
			filter:   "type=temperature,model=*",
			expected: []*Device{dev1, dev2},
		},
		{
			// devices of any type with model baz
			filter:   "type=*,model=baz",
			expected: []*Device{dev3},
		},
	}

	for _, testCase := range filterDevicesTestTable {
		actual, err := filterDevices(testCase.filter)
		expected := testCase.expected
		assert.NoError(t, err)
		assert.Equal(t, len(expected), len(actual))

		// Sort the expected and actual (we sort by model for the tests
		// since the model is unique for each test device)
		sort.SliceStable(expected, func(i int, j int) bool { return expected[i].Model < expected[j].Model })
		sort.SliceStable(actual, func(i int, j int) bool { return actual[i].Model < actual[j].Model })
		assert.Equal(t, testCase.expected, actual, "filter: %v", testCase.filter)
	}
}

// TestFilterDevicesErr tests filtering devices when the given filter
// string results in a filtering error.
func TestFilterDevicesErr(t *testing.T) {
	dev1 := &Device{
		Type:    "temperature",
		Model:   "foo",
		Handler: &DeviceHandler{},
	}
	dev2 := &Device{
		Type:    "temperature",
		Model:   "bar",
		Handler: &DeviceHandler{},
	}

	// Populate the device map with the test devices.
	deviceMap = map[string]*Device{
		"dev1": dev1,
		"dev2": dev2,
	}

	// Set up the test cases
	var filterDevicesTestTable = []string{
		// no filter - when filtering, we should always have a filter
		// string specified when calling filterDevices
		"",

		// unsupported filter keys
		"invalid=temperature",
		"MODEL=bar",
		"Type=temperature",
	}

	for _, testCase := range filterDevicesTestTable {
		_, err := filterDevices(testCase)
		assert.Error(t, err)
	}
}
