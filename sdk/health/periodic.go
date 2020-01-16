// Synse SDK
// Copyright (c) 2019-2020 Vapor IO
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
	"sync"
	"time"

	"github.com/vapor-ware/synse-sdk/sdk/utils"
)

const (
	// PeriodicCheck is the type for Checks which run periodically.
	PeriodicCheck CheckType = "periodic"
)

// PeriodicHealthCheck defines a point of health whose status can be checked periodically.
type PeriodicHealthCheck struct {
	Name  string
	Type  CheckType
	Check func() error

	interval   time.Duration
	lock       sync.RWMutex
	lastUpdate string
	err        error
}

// NewPeriodicHealthCheck creates a new health hCheck of the "periodic" type.
func NewPeriodicHealthCheck(name string, interval time.Duration, check func() error) *PeriodicHealthCheck {
	return &PeriodicHealthCheck{
		Name:     name,
		Type:     PeriodicCheck,
		Check:    check,
		interval: interval,
		lock:     sync.RWMutex{},
	}
}

// GetName gets the name of the periodic health check.
func (check *PeriodicHealthCheck) GetName() string {
	return check.Name
}

// GetType gets the type of the periodic health check.
func (check *PeriodicHealthCheck) GetType() CheckType {
	return check.Type
}

// Status gets the status of the health Check in an asynchronous-safe manner.
func (check *PeriodicHealthCheck) Status() *Status {
	check.lock.RLock()
	defer check.lock.RUnlock()

	var message string
	if check.err != nil {
		message = check.err.Error()
	}

	return &Status{
		Name:      check.Name,
		Ok:        check.err == nil,
		Message:   message,
		Timestamp: check.lastUpdate,
		Type:      check.Type,
	}
}

// Update updates the state of the health Check with the result of of the
// Check function in an asynchronous-safe manner.
func (check *PeriodicHealthCheck) Update() {
	// Perform the health check.
	err := check.Check()

	// Update the state safely.
	check.lock.Lock()
	defer check.lock.Unlock()

	check.lastUpdate = utils.GetCurrentTime()
	check.err = err
}

// Run updates the periodic health check on a periodic timer.
func (check *PeriodicHealthCheck) Run() {
	t := time.NewTicker(check.interval)
	for {
		<-t.C
		check.Update()
	}
}
