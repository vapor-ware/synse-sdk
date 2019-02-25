// Synse SDK
// Copyright (c) 2019 Vapor IO
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

// --- tests ---

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
