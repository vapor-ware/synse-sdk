package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-server-grpc/go"

	"github.com/vapor-ware/synse-sdk/internal/test"
)

// TestLocation_Encode tests encoding an SDK Location to the gRPC
// MetaLocation model.
func TestLocation_Encode(t *testing.T) {
	var cases = []struct {
		in  Location
		out synse.MetaLocation
	}{
		{
			in: Location{
				Rack:  "rack-1",
				Board: "board-1",
			},
			out: synse.MetaLocation{
				Rack:  "rack-1",
				Board: "board-1",
			},
		},
		{
			in: Location{
				Rack:  "1",
				Board: "1",
			},
			out: synse.MetaLocation{
				Rack:  "1",
				Board: "1",
			},
		},
		{
			in: Location{
				Rack:  "",
				Board: "",
			},
			out: synse.MetaLocation{
				Rack:  "",
				Board: "",
			},
		},
	}

	for _, tc := range cases {
		err := tc.in.Validate()
		assert.NoError(t, err)

		r := tc.in.Encode()
		assert.Equal(t, tc.out, *r)
	}
}

// TestLocation_GetRack tests getting the rack info for a Location successfully.
func TestLocation_GetRack(t *testing.T) {
	var cases = []struct {
		in  Location
		out string
	}{
		{
			in: Location{
				Rack:  "rack-1",
				Board: "board-1",
			},
			out: "rack-1",
		},
		{
			in: Location{
				Rack:  "1",
				Board: "1",
			},
			out: "1",
		},
		{
			in: Location{
				Rack:  "",
				Board: "",
			},
			out: "",
		},
		{
			in: Location{
				Rack: map[interface{}]interface{}{
					"from_env": "TEST_RACK_ENV",
				},
				Board: "board-1",
			},
			out: "rack-env",
		},
	}

	// Set the env variable for the test
	test.CheckErr(t, os.Setenv("TEST_RACK_ENV", "rack-env"))
	defer func() {
		test.CheckErr(t, os.Unsetenv("TEST_RACK_ENV"))
	}()

	for _, tc := range cases {
		rack, err := tc.in.GetRack()
		assert.NoError(t, err)
		assert.Equal(t, tc.out, rack)
	}
}

// TestLocation_GetRackErr tests getting the rack info for a Location unsuccessfully.
func TestLocation_GetRackErr(t *testing.T) {
	var cases = []struct {
		in  Location
		out string
	}{
		{
			in: Location{
				Rack:  1,
				Board: "board-1",
			},
			out: "rack-1",
		},
		{
			in: Location{
				Rack:  true,
				Board: "1",
			},
			out: "1",
		},
		{
			in: Location{
				Rack: map[interface{}]interface{}{
					"invalid_key": "TEST_RACK_ENV",
				},
				Board: "board-1",
			},
		},
		{
			in: Location{
				Rack: map[interface{}]interface{}{
					"from_env": "TEST_RACK_ENV_EMPTY",
				},
				Board: "board-1",
			},
			out: "",
		},
	}

	for _, tc := range cases {
		rack, err := tc.in.GetRack()
		assert.Empty(t, rack)
		assert.Error(t, err)
	}
}

// TestLocation_Validate tests validating a Location successfully.
func TestLocation_Validate(t *testing.T) {
	var cases = []Location{
		// Location with rack as string
		{
			Rack:  "rack",
			Board: "board",
		},
		// Location with rack as correct mapping
		{
			Rack: map[interface{}]interface{}{
				"from_env": "TEST_RACK_ENV",
			},
			Board: "board",
		},
	}

	// Set the env variable for the test
	test.CheckErr(t, os.Setenv("TEST_RACK_ENV", "rack-env"))
	defer func() {
		test.CheckErr(t, os.Unsetenv("TEST_RACK_ENV"))
	}()

	for _, testCase := range cases {
		err := testCase.Validate()
		assert.NoError(t, err)
	}
}

// TestLocation_ValidateErr tests validating a Location unsuccessfully.
func TestLocation_ValidateErr(t *testing.T) {
	var cases = []Location{
		// Empty Location
		{},
		// Location with rack as invalid type (int)
		{
			Rack:  2,
			Board: "board",
		},
		// Location with rack as invalid type (list)
		{
			Rack:  []interface{}{1, 2},
			Board: "board",
		},
		// Location with rack as invalid type (bool)
		{
			Rack:  false,
			Board: "board",
		},
		// Location with rack as invalid type (nil)
		{
			Rack:  nil,
			Board: "board",
		},
		// Location with rack as interface mapping, but
		// invalid map key type (int)
		{
			Rack: map[interface{}]interface{}{
				2: "TEST_RACK_ENV",
			},
			Board: "board",
		},
		// Location with rack as interface mapping, but
		// invalid map key type (bool)
		{
			Rack: map[interface{}]interface{}{
				true: "TEST_RACK_ENV",
			},
			Board: "board",
		},
		// Location with rack as interface mapping, but
		// invalid map value type (int)
		{
			Rack: map[interface{}]interface{}{
				"from_env": 2,
			},
			Board: "board",
		},
		// Location with rack as interface mapping, but
		// invalid map value type (bool)
		{
			Rack: map[interface{}]interface{}{
				"from_env": true,
			},
			Board: "board",
		},
		// Location with rack as interface mapping, but
		// unsupported key
		{
			Rack: map[interface{}]interface{}{
				"not_supported": "value",
			},
			Board: "board",
		},
		// Location with rack as interface mapping, but
		// the specified env variable doesn't exist
		{
			Rack: map[interface{}]interface{}{
				"from_env": "TEST_INVALID_ENV_VALUE",
			},
			Board: "board",
		},
	}

	for _, testCase := range cases {
		err := testCase.Validate()
		assert.Error(t, err)
	}
}

// TestParseDeviceConfig tests parsing a device config when the config
// file does not exist.
func TestParseDeviceConfig(t *testing.T) {
	// the default directory path shouldn't exist when running tests
	_, err := ParseDeviceConfig()
	assert.Error(t, err)
}

// TestParseDeviceConfig2 tests parsing device config when the config directory
// is not a directory.
func TestParseDeviceConfig2(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "test")
	assert.NoError(t, err)
	defer func() {
		test.CheckErr(t, os.Remove(tmpfile.Name()))
	}()

	test.CheckErr(t, os.Setenv(EnvDevicePath, tmpfile.Name()))
	defer func() {
		test.CheckErr(t, os.Unsetenv(EnvDevicePath))
	}()

	_, err = ParseDeviceConfig()
	assert.Error(t, err)
}

// TestParseDeviceConfig3 tests parsing device config when no valid configs are
// in the device config directory.
func TestParseDeviceConfig3(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	defer func() {
		test.CheckErr(t, os.RemoveAll(tmpdir))
	}()

	test.CheckErr(t, os.Setenv(EnvDevicePath, tmpdir))
	defer func() {
		test.CheckErr(t, os.Unsetenv(EnvDevicePath))
	}()

	res, err := ParseDeviceConfig()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(res))
}

// TestParseDeviceConfig4 tests parsing device config when no config version
// is specified in the device config file.
func TestParseDeviceConfig4(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	defer func() {
		test.CheckErr(t, os.RemoveAll(tmpdir))
	}()

	test.CheckErr(t, os.Setenv(EnvDevicePath, tmpdir))
	defer func() {
		test.CheckErr(t, os.Unsetenv(EnvDevicePath))
	}()

	data := `locations:
  r1b1:
    rack: rack-1
    board: board-1
devices:
  - type: airflow
    model: air8884
    instances:
      - id: 1
        location: r1b1
        comment: first emulated airflow device`

	tmpf := filepath.Join(tmpdir, "tmpfile.yml")
	err = ioutil.WriteFile(tmpf, []byte(data), 0666)
	assert.NoError(t, err)

	_, err = ParseDeviceConfig()
	assert.Error(t, err)
}

// TestParseDeviceConfig5 tests parsing device config when there is no handler
// defined for the specified config version.
func TestParseDeviceConfig5(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	defer func() {
		test.CheckErr(t, os.RemoveAll(tmpdir))
	}()

	test.CheckErr(t, os.Setenv(EnvDevicePath, tmpdir))
	defer func() {
		test.CheckErr(t, os.Unsetenv(EnvDevicePath))
	}()

	data := `version: 9999.9999
locations:
  r1b1:
    rack: rack-1
    board: board-1
devices:
  - type: airflow
    model: air8884
    instances:
      - id: 1
        location: r1b1
        comment: first emulated airflow device`

	tmpf := filepath.Join(tmpdir, "tmpfile.yml")
	err = ioutil.WriteFile(tmpf, []byte(data), 0666)
	assert.NoError(t, err)

	_, err = ParseDeviceConfig()
	assert.Error(t, err)
}

// TestParseDeviceConfig6 tests parsing device config when unable to
// process the config via handler.
func TestParseDeviceConfig6(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	defer func() {
		test.CheckErr(t, os.RemoveAll(tmpdir))
	}()

	test.CheckErr(t, os.Setenv(EnvDevicePath, tmpdir))
	defer func() {
		test.CheckErr(t, os.Unsetenv(EnvDevicePath))
	}()

	data := `version: 1.0
devices:
  - type: airflow
    model: air8884
    instances:
      - id: 1
        comment: first emulated airflow device`

	tmpf := filepath.Join(tmpdir, "tmpfile.yml")
	err = ioutil.WriteFile(tmpf, []byte(data), 0666)
	assert.NoError(t, err)

	_, err = ParseDeviceConfig()
	assert.Error(t, err)
}

// TestParseDeviceConfig7 tests parsing device configs successfully.
func TestParseDeviceConfig7(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	defer func() {
		test.CheckErr(t, os.RemoveAll(tmpdir))
	}()

	test.CheckErr(t, os.Setenv(EnvDevicePath, tmpdir))
	defer func() {
		test.CheckErr(t, os.Unsetenv(EnvDevicePath))
	}()

	data := `version: 1.0
locations:
  r1b1:
    rack: rack-1
    board: board-1
devices:
  - type: airflow
    model: air8884
    instances:
      - id: 1
        location: r1b1
        comment: first emulated airflow device`

	tmpf := filepath.Join(tmpdir, "tmpfile.yml")
	err = ioutil.WriteFile(tmpf, []byte(data), 0666)
	assert.NoError(t, err)

	res, err := ParseDeviceConfig()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res))
}

// TestParseDeviceConfig8 tests parsing device configs unsuccessfully using
// an environment variable to specify the root config directory.
func TestParseDeviceConfig8(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	defer func() {
		test.CheckErr(t, os.RemoveAll(tmpdir))
	}()

	test.CheckErr(t, os.Setenv(EnvDeviceConfig, tmpdir))
	defer func() {
		test.CheckErr(t, os.Unsetenv(EnvDeviceConfig))
	}()

	data := `version: 1.0
locations:
  r1b1:
    rack: rack-1
    board: board-1
devices:
  - type: airflow
    model: air8884
    instances:
      - id: 1
        location: r1b1
        comment: first emulated airflow device`

	tmpf := filepath.Join(tmpdir, "tmpfile.yml")
	err = ioutil.WriteFile(tmpf, []byte(data), 0666)
	assert.NoError(t, err)

	_, err = ParseDeviceConfig()
	assert.Error(t, err)
}

// TestParseDeviceConfig9 tests parsing device configs successfully using
// an environment variable to specify the root config directory.
func TestParseDeviceConfig9(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	defer func() {
		test.CheckErr(t, os.RemoveAll(tmpdir))
	}()

	test.CheckErr(t, os.Setenv(EnvDeviceConfig, tmpdir))
	defer func() {
		test.CheckErr(t, os.Unsetenv(EnvDeviceConfig))
	}()

	deviceDir := filepath.Join(tmpdir, "device")
	err = os.Mkdir(deviceDir, 0700)
	assert.NoError(t, err)

	data := `version: 1.0
locations:
  r1b1:
    rack: rack-1
    board: board-1
devices:
  - type: airflow
    model: air8884
    instances:
      - id: 1
        location: r1b1
        comment: first emulated airflow device`

	tmpf := filepath.Join(deviceDir, "tmpfile.yml")
	err = ioutil.WriteFile(tmpf, []byte(data), 0666)
	assert.NoError(t, err)

	res, err := ParseDeviceConfig()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res))
}

// TestParseDeviceConfig10 tests parsing device configs successfully using
// 'from_env' as the rack value.
func TestParseDeviceConfig10(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	defer func() {
		test.CheckErr(t, os.RemoveAll(tmpdir))
	}()

	test.CheckErr(t, os.Setenv(EnvDevicePath, tmpdir))
	defer func() {
		test.CheckErr(t, os.Unsetenv(EnvDevicePath))
	}()

	test.CheckErr(t, os.Setenv("SYNSE_ENV_TEST", "test-rack"))
	defer func() {
		test.CheckErr(t, os.Unsetenv("SYNSE_ENV_TEST"))
	}()

	data := `version: 1.0
locations:
  r1b1:
    rack:
      from_env: SYNSE_ENV_TEST
    board: board-1
devices:
  - type: airflow
    model: air8884
    instances:
      - id: 1
        location: r1b1
        comment: first emulated airflow device`

	tmpf := filepath.Join(tmpdir, "tmpfile.yml")
	err = ioutil.WriteFile(tmpf, []byte(data), 0666)
	assert.NoError(t, err)

	res, err := ParseDeviceConfig()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res))

	config := res[0]
	rack, err := config.Location.GetRack()
	assert.NoError(t, err)
	assert.Equal(t, "test-rack", rack)
}

// TestParseDeviceConfig11 tests parsing device configs unsuccessfully specifying
// 'from_env' as the rack value, but not having the environment variable set.
func TestParseDeviceConfig11(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	defer func() {
		test.CheckErr(t, os.RemoveAll(tmpdir))
	}()

	test.CheckErr(t, os.Setenv(EnvDevicePath, tmpdir))
	defer func() {
		test.CheckErr(t, os.Unsetenv(EnvDevicePath))
	}()

	data := `version: 1.0
locations:
  r1b1:
    rack:
      from_env: SYNSE_ENV_TEST
    board: board-1
devices:
  - type: airflow
    model: air8884
    instances:
      - id: 1
        location: r1b1
        comment: first emulated airflow device`

	tmpf := filepath.Join(tmpdir, "tmpfile.yml")
	err = ioutil.WriteFile(tmpf, []byte(data), 0666)
	assert.NoError(t, err)

	_, err = ParseDeviceConfig()
	assert.Error(t, err)
}
