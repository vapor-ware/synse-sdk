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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-server-grpc/go"
)

// TestHealthInit tests that the init function initialized things correctly.
func TestHealthInit(t *testing.T) {
	assert.NotNil(t, DefaultCatalog)
}

// TestNewCatalog tests creating a new catalog.
func TestNewCatalog(t *testing.T) {
	catalog := NewCatalog()
	assert.NotNil(t, catalog)
	assert.Empty(t, catalog.checks)
}

// TestCatalog_GetStatus tests getting the status when there are no health checks.
func TestCatalog_GetStatus(t *testing.T) {
	catalog := NewCatalog()
	statuses := catalog.GetStatus()
	assert.Empty(t, statuses)
}

// TestCatalog_GetStatus2 tests getting the status when there are health checks.
func TestCatalog_GetStatus2(t *testing.T) {
	catalog := NewCatalog()
	catalog.checks["foo"] = NewChecker("test")
	statuses := catalog.GetStatus()
	assert.Equal(t, 1, len(statuses))
}

// TestGetStatus tests getting the status from the default catalog.
func TestGetStatus(t *testing.T) {
	statuses := GetStatus()
	assert.Empty(t, statuses)
}

// TestCatalog_Register tests registering a checker with a catalog.
func TestCatalog_Register(t *testing.T) {
	catalog := NewCatalog()

	assert.Equal(t, 0, len(catalog.checks))
	catalog.Register("foo", NewChecker("test"))
	assert.Equal(t, 1, len(catalog.checks))
}

// TestRegister tests registering a checker with the default catalog.
func TestRegister(t *testing.T) {
	defer func() {
		DefaultCatalog = NewCatalog()
	}()
	assert.Equal(t, 0, len(DefaultCatalog.checks))
	Register("foo", NewChecker("test"))
	assert.Equal(t, 1, len(DefaultCatalog.checks))
}

// TestCatalog_RegisterPeriodicCheck tests registering a periodic checker with the catalog.
func TestCatalog_RegisterPeriodicCheck(t *testing.T) {
	catalog := NewCatalog()

	assert.Equal(t, 0, len(catalog.checks))
	catalog.RegisterPeriodicCheck("foo", 2*time.Second, func() error {
		return nil
	})
	assert.Equal(t, 1, len(catalog.checks))
}

// TestRegisterPeriodicCheck tests registering a periodic checker with the default catalog.
func TestRegisterPeriodicCheck(t *testing.T) {
	defer func() {
		DefaultCatalog = NewCatalog()
	}()

	assert.Equal(t, 0, len(DefaultCatalog.checks))
	RegisterPeriodicCheck("foo", 2*time.Second, func() error {
		return nil
	})
	assert.Equal(t, 1, len(DefaultCatalog.checks))
}

// TestStatus_Encode tests converting the status to its grpc message.
func TestStatus_Encode(t *testing.T) {
	s := Status{
		Name:      "foo",
		Ok:        true,
		Message:   "",
		Timestamp: "now",
		Type:      "test",
	}
	out := s.Encode()
	assert.Equal(t, "foo", out.Name)
	assert.Equal(t, synse.HealthStatus_OK, out.Status)
	assert.Equal(t, "", out.Message)
	assert.Equal(t, "now", out.Timestamp)
	assert.Equal(t, "test", out.Type)
}

// TestNewChecker tests getting a new instance of a checker.
func TestNewChecker(t *testing.T) {
	checker := NewChecker("foo")
	assert.NotNil(t, checker)
}

// TestChecker_Get tests getting the error state of the checker.
func TestChecker_Get(t *testing.T) {
	checker := NewChecker("foo")
	err := checker.Get()
	assert.NoError(t, err)
}

// TestChecker_Status tests getting the current status of the checker.
func TestChecker_Status(t *testing.T) {
	checker := NewChecker("foo")
	status := checker.Status()
	assert.Equal(t, "", status.Name)
	assert.Equal(t, true, status.Ok)
	assert.Equal(t, "", status.Message)
	assert.Equal(t, "", status.Timestamp)
	assert.Equal(t, "foo", status.Type)
}

// TestChecker_Update tests updating the checker state.
func TestChecker_Update(t *testing.T) {
	checker := NewChecker("foo")
	checker.Update(fmt.Errorf("test error message"))
	status := checker.Status()
	assert.Equal(t, "", status.Name)
	assert.Equal(t, false, status.Ok)
	assert.Equal(t, "test error message", status.Message)
	assert.NotEmpty(t, status.Timestamp)
	assert.Equal(t, "foo", status.Type)
}
