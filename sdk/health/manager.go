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
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/utils"
)

// Manager registers and runs health checks, providing a status of overall
// plugin health.
type Manager struct {
	config   *config.HealthSettings
	checks   map[string]Check
	defaults []Check
}

// NewManager creates a new instance of the health Manager component.
func NewManager(conf *config.HealthSettings, defaults ...Check) *Manager {
	if conf == nil {
		panic("manager cannot be initialized with nil config")
	}

	return &Manager{
		config:   conf,
		checks:   make(map[string]Check),
		defaults: defaults,
	}
}

// Register registers a health check with the manager. If the name of the Check
// being registered conflicts with an existing Check's name, an error is returned.
func (manager *Manager) Register(check Check) error {
	if check.GetName() == "" {
		// fixme: better err
		return fmt.Errorf("check must have name")
	}
	if _, exists := manager.checks[check.GetName()]; exists {
		// fixme: better err
		return fmt.Errorf("check with name already exists")
	}

	// todo: log
	manager.checks[check.GetName()] = check
	return nil
}

// RegisterDefault registers default health checks with the health manager.
func (manager *Manager) RegisterDefault(check Check) {
	log.WithFields(log.Fields{
		"name": check.GetName(),
		"type": check.GetType(),
	}).Info("[health] adding default health check")
	manager.defaults = append(manager.defaults, check)
}

// Init initializes the health Manager, making sure its necessary state is set.
func (manager *Manager) Init() error {
	// If a health file is configured, ensure that the directory exists so we
	// can create the file later.
	if manager.config.HealthFile != "" {
		dir, _ := filepath.Split(manager.config.HealthFile)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

// Start starts the health Manager.
func (manager *Manager) Start() {
	// Run default health checks, if enabled.
	if !manager.config.Checks.DisableDefaults {
		for _, check := range manager.defaults {
			go check.Run()
		}
	}

	// Run any custom health checks.
	for _, check := range manager.checks {
		go check.Run()
	}

	// Update the health file, if configured.
	if manager.config.HealthFile != "" {
		go func() {
			t := time.NewTicker(manager.config.UpdateInterval)
			for {
				<-t.C
				if err := manager.updateHealthFile(); err != nil {
					log.WithField("error", err).Errorf("[health] failed to update health file")
				}
			}
		}()
	}
}

// Status gets a current summary of plugin health based on the configured health checks.
func (manager *Manager) Status() *Summary {
	var checks []*Status
	var ok = true

	// Get the status of default checks, if enabled.
	if !manager.config.Checks.DisableDefaults {
		for _, check := range manager.defaults {
			s := check.Status()
			if !s.Ok {
				ok = false
			}
			checks = append(checks, s)
		}
	}

	// Get the status of any custom checks.
	for _, check := range manager.checks {
		s := check.Status()
		if !s.Ok {
			ok = false
		}
		checks = append(checks, s)
	}

	return &Summary{
		Timestamp: utils.GetCurrentTime(),
		Ok:        ok,
		Checks:    checks,
	}
}

// updateHealthFile updates the health file (by either adding or removing it) based
// on the current plugin health status.
func (manager *Manager) updateHealthFile() error {
	if manager.config.HealthFile == "" {
		return fmt.Errorf("cannot update health file, no health file defined")
	}

	// First, get the health summary.
	summary := manager.Status()

	// Determine if the health file already exists.
	var exists bool
	if _, err := os.Stat(manager.config.HealthFile); err == nil {
		exists = true
	} else if os.IsNotExist(err) {
		exists = false
	} else {
		return err
	}

	if summary.Ok && !exists {
		// The status is OK and the health file is not present; add it.
		if err := ioutil.WriteFile(manager.config.HealthFile, []byte("ok"), os.ModePerm); err != nil {
			return err
		}
	} else if !summary.Ok && exists {
		// The status is not OK and the health file exists; remove it.
		if err := os.Remove(manager.config.HealthFile); err != nil {
			return err
		}
	}

	return nil
}
