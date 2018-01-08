package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/vapor-ware/synse-server-grpc/go"
)

var deviceOutputEncodeTestTable = []struct {
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

func TestDeviceOutput_Encode(t *testing.T) {
	for i, tc := range deviceOutputEncodeTestTable {
		r := tc.in.Encode()
		if r.Type != tc.out.Type {
			t.Errorf("(%d) %v.Encode => %v, want %v", i, tc.in, r, tc.out)
		}
		if r.DataType != tc.out.DataType {
			t.Errorf("(%d) %v.Encode => %v, want %v", i, tc.in, r, tc.out)
		}
		if r.Precision != tc.out.Precision {
			t.Errorf("(%d) %v.Encode => %v, want %v", i, tc.in, r, tc.out)
		}
		if *r.Unit != *tc.out.Unit {
			t.Errorf("(%d) %v.Encode => %v, want %v", i, tc.in, r, tc.out)
		}
		if *r.Range != *tc.out.Range {
			t.Errorf("(%d) %v.Encode => %v, want %v", i, tc.in, r, tc.out)
		}
	}
}

var unitEncodeTestTable = []struct {
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

func TestUnit_Encode(t *testing.T) {
	for _, tc := range unitEncodeTestTable {
		r := tc.in.Encode()
		if *r != tc.out {
			t.Errorf("%v.Encode() => %v, want %v", tc.in, r, tc.out)
		}
	}
}

var rangeEncodeTestTable = []struct {
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

func TestRange_Encode(t *testing.T) {
	for _, tc := range rangeEncodeTestTable {
		r := tc.in.Encode()
		if *r != tc.out {
			t.Errorf("%v.Encode() => %v, want %v", tc.in, r, tc.out)
		}
	}
}

// config file doesn't exist
func TestParseProtoConfig(t *testing.T) {
	// the default directory path shouldn't exist when running tests
	_, err := ParsePrototypeConfig()
	if err == nil {
		t.Error("expected error: config directory should not exist")
	}
}

// config directory is not a directory
func TestParseProtoConfig2(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmpfile.Name())

	os.Setenv(EnvProtoPath, tmpfile.Name())
	defer os.Unsetenv(EnvProtoPath)

	_, err = ParsePrototypeConfig()
	if err == nil {
		t.Error("expected error: config directory should not be a directory")
	}
}

// no valid configs in directory
func TestParseProtoConfig3(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(tmpdir)

	os.Setenv(EnvProtoPath, tmpdir)
	defer os.Unsetenv(EnvProtoPath)

	res, err := ParsePrototypeConfig()
	if err != nil {
		t.Error(err)
	}
	if len(res) > 0 {
		t.Errorf("expected 0 results, but got %v instead", len(res))
	}
}

// no config version specified
func TestParseProtoConfig4(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Error(err)
	}
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
	if err != nil {
		t.Error(err)
	}

	_, err = ParsePrototypeConfig()
	if err == nil {
		t.Error("expected error: configuration version not set")
	}
}

// no handler for the specified config version
func TestParseProtoConfig5(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Error(err)
	}
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
	if err != nil {
		t.Error(err)
	}

	_, err = ParsePrototypeConfig()
	if err == nil {
		t.Error("expected error: no handler for the given version")
	}
}

// unable to process the config via handler
func TestParseProtoConfig6(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(tmpdir)

	os.Setenv(EnvProtoPath, tmpdir)
	defer os.Unsetenv(EnvProtoPath)

	data := `version: 1.0`

	tmpf := filepath.Join(tmpdir, "tmpfile.yml")
	err = ioutil.WriteFile(tmpf, []byte(data), 0666)
	if err != nil {
		t.Error(err)
	}

	_, err = ParsePrototypeConfig()
	if err == nil {
		t.Error("expected error: invalid config for given version")
	}
}

// process everything successfully
func TestParseProtoConfig7(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Error(err)
	}
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
	if err != nil {
		t.Error(err)
	}

	res, err := ParsePrototypeConfig()
	if err != nil {
		t.Error(err)
	}
	if len(res) != 1 {
		t.Errorf("expected 1 prototype configuration, but got %v", len(res))
	}
}
