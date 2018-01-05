package config

import (
	"github.com/vapor-ware/synse-server-grpc/go"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var locationEncodeTestTable = []struct {
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

func TestLocation_Encode(t *testing.T) {
	for _, tc := range locationEncodeTestTable {
		r := tc.in.Encode()
		if *r != tc.out {
			t.Errorf("%v.Encode() => %v, want %v", tc.in, r, tc.out)
		}
	}
}

// config file doesn't exist
func TestParseDeviceConfig(t *testing.T) {
	// the default directory path shouldn't exist when running tests
	_, err := ParseDeviceConfig()
	if err == nil {
		t.Error("expected error: config directory should not exist")
	}
}

// config directory is not a directory
func TestParseDeviceConfig2(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmpfile.Name())

	os.Setenv(EnvDevicePath, tmpfile.Name())
	defer os.Unsetenv(EnvDevicePath)

	_, err = ParseDeviceConfig()
	if err == nil {
		t.Error("expected error: config directory should not be a directory")
	}
}

// no valid configs in directory
func TestParseDeviceConfig3(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(tmpdir)

	os.Setenv(EnvDevicePath, tmpdir)
	defer os.Unsetenv(EnvDevicePath)

	res, err := ParseDeviceConfig()
	if err != nil {
		t.Error(err)
	}
	if len(res) > 0 {
		t.Errorf("expected 0 results, but got %v instead", len(res))
	}
}

// no config version specified
func TestParseDeviceConfig4(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(tmpdir)

	os.Setenv(EnvDevicePath, tmpdir)
	defer os.Unsetenv(EnvDevicePath)

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
	if err != nil {
		t.Error(err)
	}

	_, err = ParseDeviceConfig()
	if err == nil {
		t.Error("expected error: configuration version not set")
	}
}

// no handler for the specified config version
func TestParseDeviceConfig5(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(tmpdir)

	os.Setenv(EnvDevicePath, tmpdir)
	defer os.Unsetenv(EnvDevicePath)

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
	if err != nil {
		t.Error(err)
	}

	_, err = ParseDeviceConfig()
	if err == nil {
		t.Error("expected error: no handler for the given version")
	}
}

// unable to process the config via handler
func TestParseDeviceConfig6(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(tmpdir)

	os.Setenv(EnvDevicePath, tmpdir)
	defer os.Unsetenv(EnvDevicePath)

	data := `version: 1.0
devices:
  - type: airflow
    model: air8884
    instances:
      - id: 1
        comment: first emulated airflow device`

	tmpf := filepath.Join(tmpdir, "tmpfile.yml")
	err = ioutil.WriteFile(tmpf, []byte(data), 0666)
	if err != nil {
		t.Error(err)
	}

	_, err = ParseDeviceConfig()
	if err == nil {
		t.Error("expected error: invalid config for given version")
	}
}

// process everything successfully
func TestParseDeviceConfig7(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(tmpdir)

	os.Setenv(EnvDevicePath, tmpdir)
	defer os.Unsetenv(EnvDevicePath)

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
	if err != nil {
		t.Error(err)
	}

	res, err := ParseDeviceConfig()
	if err != nil {
		t.Error(err)
	}
	if len(res) != 1 {
		t.Errorf("expected 1 device configuration, but got %v", len(res))
	}
}
