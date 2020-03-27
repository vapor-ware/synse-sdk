// Synse SDK
// Copyright (c) 2017-2020 Vapor IO
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package health

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/internal/test"
	"github.com/vapor-ware/synse-sdk/sdk/config"
)

//
// define test health check
//

type testCheck struct {
	ok   bool
	name string
}

func (check *testCheck) GetName() string {
	return check.name
}

func (check *testCheck) GetType() CheckType {
	return PeriodicCheck
}

func (check *testCheck) Status() *Status {
	return &Status{
		Name:      check.name,
		Ok:        check.ok,
		Message:   "foo",
		Timestamp: "now",
		Type:      PeriodicCheck,
	}
}

func (check *testCheck) Update() {}
func (check *testCheck) Run()    {}

// --- test cases ---

func TestNewManager(t *testing.T) {
	conf := &config.HealthSettings{
		HealthFile:     "/tmp",
		UpdateInterval: 30 * time.Second,
	}
	m := NewManager(conf)

	assert.Equal(t, conf, m.config)
	assert.Empty(t, m.checks)
	assert.Empty(t, m.defaults)
}

func TestNewManager_nilConfig(t *testing.T) {
	assert.Panics(t, func() {
		NewManager(nil)
	})
}

func TestManager_Count(t *testing.T) {
	// no checks
	m := Manager{
		checks:   map[string]Check{},
		defaults: []Check{},
	}

	assert.Equal(t, 0, m.Count())
}

func TestManager_Count2(t *testing.T) {
	// only default checks
	m := Manager{
		checks: map[string]Check{},
		defaults: []Check{
			&testCheck{name: "foo"},
		},
	}

	assert.Equal(t, 1, m.Count())
}

func TestManager_Count3(t *testing.T) {
	// only custom checks
	m := Manager{
		checks: map[string]Check{
			"foo": &testCheck{name: "foo"},
		},
		defaults: []Check{},
	}

	assert.Equal(t, 1, m.Count())
}

func TestManager_Count4(t *testing.T) {
	// custom and default checks
	m := Manager{
		checks: map[string]Check{
			"foo": &testCheck{name: "foo"},
		},
		defaults: []Check{
			&testCheck{name: "bar"},
		},
	}

	assert.Equal(t, 2, m.Count())
}

func TestManager_Register(t *testing.T) {
	check := testCheck{
		name: "foo",
	}
	m := Manager{
		checks: make(map[string]Check),
	}

	assert.Len(t, m.checks, 0)
	err := m.Register(&check)
	assert.NoError(t, err)
	assert.Len(t, m.checks, 1)
}

func TestManager_Register_noName(t *testing.T) {
	check := testCheck{}
	m := Manager{
		checks: make(map[string]Check),
	}

	err := m.Register(&check)
	assert.Error(t, err)
	assert.Empty(t, m.checks)
}

func TestManager_Register_alreadyExists(t *testing.T) {
	check := testCheck{
		name: "foo",
	}
	m := Manager{
		checks: make(map[string]Check),
	}
	m.checks["foo"] = &testCheck{}

	assert.Len(t, m.checks, 1)
	err := m.Register(&check)
	assert.Error(t, err)
	assert.Len(t, m.checks, 1)
}

func TestManager_RegisterDefault(t *testing.T) {
	check := testCheck{}
	m := Manager{}

	assert.Empty(t, m.defaults)
	m.RegisterDefault(&check)
	assert.Len(t, m.defaults, 1)
}

func TestManager_Init(t *testing.T) {
	d, closer := test.TempDir(t)
	defer closer()

	// define the health file
	healthFile := filepath.Join(d, "health")

	m := Manager{
		config: &config.HealthSettings{
			HealthFile: healthFile,
		},
	}

	// health file directory already exists
	err := m.Init()
	assert.NoError(t, err)
}

func TestManager_Init2(t *testing.T) {
	d, closer := test.TempDir(t)
	defer closer()

	// define the health file
	healthDir := filepath.Join(d, "foo", "bar")
	healthFile := filepath.Join(healthDir, "health")

	m := Manager{
		config: &config.HealthSettings{
			HealthFile: healthFile,
		},
	}

	// health file directory does not already exist
	_, err := os.Stat(healthDir)
	assert.Error(t, err)
	assert.IsType(t, &os.PathError{}, err)

	err = m.Init()
	assert.NoError(t, err)

	_, err = os.Stat(healthDir)
	assert.NoError(t, err)
}

func TestManager_Status(t *testing.T) {
	// Set up the manager with defaults enabled, but no defaults set
	// and no other checks set.
	m := Manager{
		config: &config.HealthSettings{
			HealthFile: "",
			Checks: &config.HealthCheckSettings{
				DisableDefaults: false,
			},
		},
	}

	status := m.Status()
	assert.NotEmpty(t, status.Timestamp)
	assert.True(t, status.Ok)
	assert.Len(t, status.Checks, 0)
}

func TestManager_Status2(t *testing.T) {
	// Set up the manager with defaults enabled, and defaults set
	// and no other checks set.
	m := Manager{
		config: &config.HealthSettings{
			HealthFile: "",
			Checks: &config.HealthCheckSettings{
				DisableDefaults: false,
			},
		},
		defaults: []Check{
			&testCheck{ok: true, name: "check1"},
		},
	}

	status := m.Status()
	assert.NotEmpty(t, status.Timestamp)
	assert.True(t, status.Ok)
	assert.Len(t, status.Checks, 1)
}

func TestManager_Status3(t *testing.T) {
	// Set up the manager with defaults disabled, defaults set,
	// and no other checks set.
	m := Manager{
		config: &config.HealthSettings{
			HealthFile: "",
			Checks: &config.HealthCheckSettings{
				DisableDefaults: true,
			},
		},
		defaults: []Check{
			&testCheck{ok: true, name: "check1"},
		},
	}

	status := m.Status()
	assert.NotEmpty(t, status.Timestamp)
	assert.True(t, status.Ok)
	assert.Len(t, status.Checks, 0)
}

func TestManager_Status4(t *testing.T) {
	// Set up the manager with defaults disabled, but no defaults set
	// and no other checks set.
	m := Manager{
		config: &config.HealthSettings{
			HealthFile: "",
			Checks: &config.HealthCheckSettings{
				DisableDefaults: true,
			},
		},
	}

	status := m.Status()
	assert.NotEmpty(t, status.Timestamp)
	assert.True(t, status.Ok)
	assert.Len(t, status.Checks, 0)
}

func TestManager_Status5(t *testing.T) {
	// Set up the manager with no defaults, only custom checks.
	m := Manager{
		config: &config.HealthSettings{
			HealthFile: "",
			Checks: &config.HealthCheckSettings{
				DisableDefaults: false,
			},
		},
		checks: map[string]Check{
			"check1": &testCheck{ok: true, name: "check1"},
		},
	}

	status := m.Status()
	assert.NotEmpty(t, status.Timestamp)
	assert.True(t, status.Ok)
	assert.Len(t, status.Checks, 1)
}

func TestManager_Status6(t *testing.T) {
	// Get the status when all checks fail.
	m := Manager{
		config: &config.HealthSettings{
			HealthFile: "",
			Checks: &config.HealthCheckSettings{
				DisableDefaults: false,
			},
		},
		defaults: []Check{
			&testCheck{ok: false, name: "check1"},
			&testCheck{ok: false, name: "check2"},
		},
		checks: map[string]Check{
			"check3": &testCheck{ok: false, name: "check3"},
			"check4": &testCheck{ok: false, name: "check4"},
		},
	}

	status := m.Status()
	assert.NotEmpty(t, status.Timestamp)
	assert.False(t, status.Ok)
	assert.Len(t, status.Checks, 4)
}

func TestManager_Status7(t *testing.T) {
	// Get the status when only one default check is failing.
	m := Manager{
		config: &config.HealthSettings{
			HealthFile: "",
			Checks: &config.HealthCheckSettings{
				DisableDefaults: false,
			},
		},
		defaults: []Check{
			&testCheck{ok: true, name: "check1"},
			&testCheck{ok: false, name: "check2"},
		},
		checks: map[string]Check{
			"check3": &testCheck{ok: true, name: "check3"},
			"check4": &testCheck{ok: true, name: "check4"},
		},
	}

	status := m.Status()
	assert.NotEmpty(t, status.Timestamp)
	assert.False(t, status.Ok)
	assert.Len(t, status.Checks, 4)
}

func TestManager_Status8(t *testing.T) {
	// Get the status when only one custom check is failing.
	m := Manager{
		config: &config.HealthSettings{
			HealthFile: "",
			Checks: &config.HealthCheckSettings{
				DisableDefaults: false,
			},
		},
		defaults: []Check{
			&testCheck{ok: true, name: "check1"},
			&testCheck{ok: true, name: "check2"},
		},
		checks: map[string]Check{
			"check3": &testCheck{ok: false, name: "check3"},
			"check4": &testCheck{ok: true, name: "check4"},
		},
	}

	status := m.Status()
	assert.NotEmpty(t, status.Timestamp)
	assert.False(t, status.Ok)
	assert.Len(t, status.Checks, 4)
}

func TestManager_updateHealthFile(t *testing.T) {
	// No health file configured.
	m := Manager{
		config: &config.HealthSettings{
			HealthFile: "",
		},
	}

	err := m.updateHealthFile()
	assert.Error(t, err)
}

func TestManager_updateHealthFile2(t *testing.T) {
	// Health file configured, status is healthy, file does not already exist.
	d, closer := test.TempDir(t)
	defer closer()

	healthFile := filepath.Join(d, "health")

	m := Manager{
		config: &config.HealthSettings{
			HealthFile: healthFile,
			Checks: &config.HealthCheckSettings{
				DisableDefaults: false,
			},
		},
		checks: map[string]Check{
			"check1": &testCheck{ok: true, name: "check1"},
		},
	}

	// health file does not exist
	_, err := os.Stat(healthFile)
	assert.Error(t, err)
	assert.IsType(t, &os.PathError{}, err)

	err = m.updateHealthFile()
	assert.NoError(t, err)

	// health file exists
	_, err = os.Stat(healthFile)
	assert.NoError(t, err)
}

func TestManager_updateHealthFile3(t *testing.T) {
	// Health file configured, status is healthy, file already exists.
	d, closer := test.TempDir(t)
	defer closer()

	healthFile := filepath.Join(d, "health")
	err := ioutil.WriteFile(healthFile, []byte("ok"), os.ModePerm)
	assert.NoError(t, err)

	m := Manager{
		config: &config.HealthSettings{
			HealthFile: healthFile,
			Checks: &config.HealthCheckSettings{
				DisableDefaults: false,
			},
		},
		checks: map[string]Check{
			"check1": &testCheck{ok: true, name: "check1"},
		},
	}

	// health file exists
	_, err = os.Stat(healthFile)
	assert.NoError(t, err)

	err = m.updateHealthFile()
	assert.NoError(t, err)

	// health file exists
	_, err = os.Stat(healthFile)
	assert.NoError(t, err)
}

func TestManager_updateHealthFile4(t *testing.T) {
	// Health file configured, status is unhealthy, file does not exist.
	d, closer := test.TempDir(t)
	defer closer()

	healthFile := filepath.Join(d, "health")

	m := Manager{
		config: &config.HealthSettings{
			HealthFile: healthFile,
			Checks: &config.HealthCheckSettings{
				DisableDefaults: false,
			},
		},
		checks: map[string]Check{
			"check1": &testCheck{ok: false, name: "check1"},
		},
	}

	// health file does not exist
	_, err := os.Stat(healthFile)
	assert.Error(t, err)
	assert.IsType(t, &os.PathError{}, err)

	err = m.updateHealthFile()
	assert.NoError(t, err)

	// health file does not exist
	_, err = os.Stat(healthFile)
	assert.Error(t, err)
	assert.IsType(t, &os.PathError{}, err)
}

func TestManager_updateHealthFile5(t *testing.T) {
	// Health file configured, status is unhealthy, file does exist.
	d, closer := test.TempDir(t)
	defer closer()

	healthFile := filepath.Join(d, "health")
	err := ioutil.WriteFile(healthFile, []byte("ok"), os.ModePerm)
	assert.NoError(t, err)

	m := Manager{
		config: &config.HealthSettings{
			HealthFile: healthFile,
			Checks: &config.HealthCheckSettings{
				DisableDefaults: false,
			},
		},
		checks: map[string]Check{
			"check1": &testCheck{ok: false, name: "check1"},
		},
	}

	// health file exists
	_, err = os.Stat(healthFile)
	assert.NoError(t, err)

	err = m.updateHealthFile()
	assert.NoError(t, err)

	// health file does not exist
	_, err = os.Stat(healthFile)
	assert.Error(t, err)
	assert.IsType(t, &os.PathError{}, err)
}
