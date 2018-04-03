package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vapor-ware/synse-server-grpc/go"
)

// TestDeviceOutput_Encode tests encoding an SDK DeviceOutput to
// the gRPC MetaOutput model.
func TestDeviceOutput_Encode(t *testing.T) {
	var cases = []struct {
		in  DeviceOutput
		out synse.MetaOutput
	}{
		{
			in: DeviceOutput{},
			out: synse.MetaOutput{
				Unit:  &synse.MetaOutputUnit{},
				Range: &synse.MetaOutputRange{},
			},
		},
		{
			in: DeviceOutput{
				Type: "test",
			},
			out: synse.MetaOutput{
				Type:  "test",
				Unit:  &synse.MetaOutputUnit{},
				Range: &synse.MetaOutputRange{},
			},
		},
		{
			in: DeviceOutput{
				Type:      "test",
				DataType:  "test",
				Precision: 10,
			},
			out: synse.MetaOutput{
				Type:      "test",
				DataType:  "test",
				Precision: 10,
				Unit:      &synse.MetaOutputUnit{},
				Range:     &synse.MetaOutputRange{},
			},
		},
		{
			in: DeviceOutput{
				Type:      "test",
				DataType:  "test",
				Precision: 10,
				Unit: &Unit{
					Name:   "test",
					Symbol: "t",
				},
				Range: &Range{
					Min: 10,
					Max: 100,
				},
			},
			out: synse.MetaOutput{
				Type:      "test",
				DataType:  "test",
				Precision: 10,
				Unit: &synse.MetaOutputUnit{
					Name:   "test",
					Symbol: "t",
				},
				Range: &synse.MetaOutputRange{
					Min: 10,
					Max: 100,
				},
			},
		},
	}

	for _, tc := range cases {
		r := tc.in.Encode()

		assert.Equal(t, tc.out.Type, r.Type)
		assert.Equal(t, tc.out.DataType, r.DataType)
		assert.Equal(t, tc.out.Precision, r.Precision)
		assert.Equal(t, *tc.out.Unit, *r.Unit)
		assert.Equal(t, *tc.out.Range, *r.Range)
	}
}

// TestUnit_Encode tests encoding the SDK Unit to the gRPC
// MetaOutputUnit model.
func TestUnit_Encode(t *testing.T) {
	var cases = []struct {
		in  Unit
		out synse.MetaOutputUnit
	}{
		{
			in:  Unit{},
			out: synse.MetaOutputUnit{},
		},
		{
			in: Unit{
				Name: "test",
			},
			out: synse.MetaOutputUnit{
				Name: "test",
			},
		},
		{
			in: Unit{
				Name:   "test",
				Symbol: "t",
			},
			out: synse.MetaOutputUnit{
				Name:   "test",
				Symbol: "t",
			},
		},
	}

	for _, tc := range cases {
		r := tc.in.Encode()
		assert.Equal(t, tc.out, *r)
	}
}

// TestRange_Encode tests encoding the SDK Range to the gRPC
// MetaOutputRange model.
func TestRange_Encode(t *testing.T) {
	var cases = []struct {
		in  Range
		out synse.MetaOutputRange
	}{
		{
			in:  Range{},
			out: synse.MetaOutputRange{},
		},
		{
			in: Range{
				Min: 10,
			},
			out: synse.MetaOutputRange{
				Min: 10,
			},
		},
		{
			in: Range{
				Min: 10,
				Max: 100,
			},
			out: synse.MetaOutputRange{
				Min: 10,
				Max: 100,
			},
		},
	}

	for _, tc := range cases {
		r := tc.in.Encode()
		assert.Equal(t, tc.out, *r)
	}
}

// TestParseProtoConfig tests parsing prototype config when a prototype config
// file doesn't exist.
func TestParseProtoConfig(t *testing.T) {
	// the default directory path shouldn't exist when running tests
	_, err := ParsePrototypeConfig()
	assert.Error(t, err)
}

// TestParseProtoConfig2 tests parsing prototype config when the prototype config
// directory is not a directory.
func TestParseProtoConfig2(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "test")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	os.Setenv(EnvProtoPath, tmpfile.Name())
	defer os.Unsetenv(EnvProtoPath)

	_, err = ParsePrototypeConfig()
	assert.Error(t, err)
}

// TestParseProtoConfig3 tests parsing prototype config when no valid configs
// are in the prototype config directory.
func TestParseProtoConfig3(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	os.Setenv(EnvProtoPath, tmpdir)
	defer os.Unsetenv(EnvProtoPath)

	res, err := ParsePrototypeConfig()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(res))
}

// TestParseProtoConfig4 tests parsing prototype config when no config version
// is specified in the prototype config.
func TestParseProtoConfig4(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	os.Setenv(EnvProtoPath, tmpdir)
	defer os.Unsetenv(EnvProtoPath)

	data := `prototypes:
  - type: airflow
    model: air8884
    manufacturer: vaporio
    protocol: emulator
    output:
      - type: airflow
        data_type: float
        unit:
          name: cubic feet per minute
          symbol: CFM
        precision: 2
        range:
          min: 0
          max: 1000`

	tmpf := filepath.Join(tmpdir, "tmpfile.yml")
	err = ioutil.WriteFile(tmpf, []byte(data), 0666)
	assert.NoError(t, err)

	_, err = ParsePrototypeConfig()
	assert.Error(t, err)
}

// TestParseProtoConfig5 tests parsing prototype config when no handler for the
// specified config version is defined.
func TestParseProtoConfig5(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	os.Setenv(EnvProtoPath, tmpdir)
	defer os.Unsetenv(EnvProtoPath)

	data := `version: 9999.9999
prototypes:
  - type: airflow
    model: air8884
    manufacturer: vaporio
    protocol: emulator
    output:
      - type: airflow
        data_type: float
        unit:
          name: cubic feet per minute
          symbol: CFM
        precision: 2
        range:
          min: 0
          max: 1000`

	tmpf := filepath.Join(tmpdir, "tmpfile.yml")
	err = ioutil.WriteFile(tmpf, []byte(data), 0666)
	assert.NoError(t, err)

	_, err = ParsePrototypeConfig()
	assert.Error(t, err)
}

// TestParseProtoConfig6 tests parsing prototype config when unable to process
// the config via handler.
func TestParseProtoConfig6(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	os.Setenv(EnvProtoPath, tmpdir)
	defer os.Unsetenv(EnvProtoPath)

	data := `version: 1.0`

	tmpf := filepath.Join(tmpdir, "tmpfile.yml")
	err = ioutil.WriteFile(tmpf, []byte(data), 0666)
	assert.NoError(t, err)

	_, err = ParsePrototypeConfig()
	assert.Error(t, err)
}

// TestParseProtoConfig7 tests parsing the prototype configuration successfully.
func TestParseProtoConfig7(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	os.Setenv(EnvProtoPath, tmpdir)
	defer os.Unsetenv(EnvProtoPath)

	data := `version: 1.0
prototypes:
  - type: airflow
    model: air8884
    manufacturer: vaporio
    protocol: emulator
    output:
      - type: airflow
        data_type: float
        unit:
          name: cubic feet per minute
          symbol: CFM
        precision: 2
        range:
          min: 0
          max: 1000`

	tmpf := filepath.Join(tmpdir, "tmpfile.yml")
	err = ioutil.WriteFile(tmpf, []byte(data), 0666)
	assert.NoError(t, err)

	res, err := ParsePrototypeConfig()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res))
}

// TestParseProtoConfig8 tests parsing prototype configs unsuccessfully using
// an environment variable to specify the root config directory.
func TestParseProtoConfig8(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	os.Setenv(EnvDeviceConfig, tmpdir)
	defer os.Unsetenv(EnvDeviceConfig)

	data := `version: 1.0
prototypes:
  - type: airflow
    model: air8884
    manufacturer: vaporio
    protocol: emulator
    output:
      - type: airflow
        data_type: float
        unit:
          name: cubic feet per minute
          symbol: CFM
        precision: 2
        range:
          min: 0
          max: 1000`

	tmpf := filepath.Join(tmpdir, "tmpfile.yml")
	err = ioutil.WriteFile(tmpf, []byte(data), 0666)
	assert.NoError(t, err)

	_, err = ParsePrototypeConfig()
	if err == nil {
		t.Error("expected error: PLUGIN_DEVICE_CONFIG path does not contain 'proto' subdir")
	}
}

// TestParseProtoConfig9 tests parsing the prototype config successfully using
// an environment variable to specify the root config directory.
func TestParseProtoConfig9(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	os.Setenv(EnvDeviceConfig, tmpdir)
	defer os.Unsetenv(EnvDeviceConfig)

	protoDir := filepath.Join(tmpdir, "proto")
	os.Mkdir(protoDir, 0700)

	data := `version: 1.0
prototypes:
  - type: airflow
    model: air8884
    manufacturer: vaporio
    protocol: emulator
    output:
      - type: airflow
        data_type: float
        unit:
          name: cubic feet per minute
          symbol: CFM
        precision: 2
        range:
          min: 0
          max: 1000`

	tmpf := filepath.Join(protoDir, "tmpfile.yml")
	err = ioutil.WriteFile(tmpf, []byte(data), 0666)
	assert.NoError(t, err)

	res, err := ParsePrototypeConfig()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res))
}
