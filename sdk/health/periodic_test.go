// Synse SDK
// Copyright (c) 2017-2022 Vapor IO
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
)

func TestNewPeriodicHealthCheck(t *testing.T) {
	hc := NewPeriodicHealthCheck("foo", 3*time.Second, func() error { return nil })

	assert.Equal(t, "foo", hc.Name)
	assert.Equal(t, PeriodicCheck, hc.Type)
	assert.Equal(t, 3*time.Second, hc.interval)
}

func TestPeriodicHealthCheck_GetName(t *testing.T) {
	p := PeriodicHealthCheck{}
	assert.Equal(t, "", p.GetName())
}

func TestPeriodicHealthCheck_GetName2(t *testing.T) {
	p := PeriodicHealthCheck{
		Name: "test",
	}
	assert.Equal(t, "test", p.GetName())
}

func TestPeriodicHealthCheck_GetType(t *testing.T) {
	p := PeriodicHealthCheck{}
	assert.Equal(t, CheckType(""), p.GetType())
}

func TestPeriodicHealthCheck_GetType2(t *testing.T) {
	p := PeriodicHealthCheck{
		Type: PeriodicCheck,
	}
	assert.Equal(t, PeriodicCheck, p.GetType())
}

func TestPeriodicHealthCheck_Status(t *testing.T) {
	p := PeriodicHealthCheck{
		Name: "test",
		Type: PeriodicCheck,
	}

	status := p.Status()
	assert.Equal(t, "test", status.Name)
	assert.Equal(t, true, status.Ok)
	assert.Equal(t, "", status.Message)
	assert.Equal(t, "", status.Timestamp)
	assert.Equal(t, PeriodicCheck, status.Type)
}

func TestPeriodicHealthCheck_Status2(t *testing.T) {
	p := PeriodicHealthCheck{
		Name: "test",
		Type: PeriodicCheck,
		err:  fmt.Errorf("test error"),
	}

	status := p.Status()
	assert.Equal(t, "test", status.Name)
	assert.Equal(t, false, status.Ok)
	assert.Equal(t, "test error", status.Message)
	assert.Equal(t, "", status.Timestamp)
	assert.Equal(t, PeriodicCheck, status.Type)
}

func TestPeriodicHealthCheck_Update_ok(t *testing.T) {
	p := PeriodicHealthCheck{
		Name:  "test",
		Type:  PeriodicCheck,
		Check: func() error { return nil },
	}

	assert.Equal(t, "test", p.Name)
	assert.Equal(t, PeriodicCheck, p.Type)
	assert.Nil(t, p.err)
	assert.Empty(t, p.lastUpdate)

	p.Update()

	assert.Equal(t, "test", p.Name)
	assert.Equal(t, PeriodicCheck, p.Type)
	assert.Nil(t, p.err)
	assert.NotEmpty(t, p.lastUpdate)
}

func TestPeriodicHealthCheck_Update_err(t *testing.T) {
	p := PeriodicHealthCheck{
		Name:  "test",
		Type:  PeriodicCheck,
		Check: func() error { return fmt.Errorf("test error") },
	}

	assert.Equal(t, "test", p.Name)
	assert.Equal(t, PeriodicCheck, p.Type)
	assert.Nil(t, p.err)
	assert.Empty(t, p.lastUpdate)

	p.Update()

	assert.Equal(t, "test", p.Name)
	assert.Equal(t, PeriodicCheck, p.Type)
	assert.NotNil(t, p.err)
	assert.NotEmpty(t, p.lastUpdate)
}
