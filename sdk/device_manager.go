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

package sdk

import (
	"fmt"
	cfg "github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	log "github.com/Sirupsen/logrus"
)

const (
	// DeviceEnvOverride defines the environment variable that can be used to
	// set an override config location for device configuration files.
	DeviceEnvOverride = "PLUGIN_DEVICE_CONFIG"
)

var DeviceManager *deviceManager

func init() {
	DeviceManager = NewDeviceManager()
}

func GetDeviceByID() {}
func GetDeviceByAlias() {}
func GetDevices() {}
func GetDevicesForHandler() {}


// todo: figure out where dynamic device registration fits in here.


// DeviceAction defines an action that can be run before the main Plugin run
// logic. This is generally used for doing device-specific setup actions.
type DeviceAction struct {
	Name string
	Action func(p *Plugin, d *Device) error
}

// DeviceManager loads and manages a Plugin's devices.
type deviceManager struct {
	config  *cfg.Devices

	devices []*Device

	setupActions map[string][]*DeviceAction
}

// NewDeviceManager creates a new DeviceManager.
func NewDeviceManager() *deviceManager {
	return &deviceManager{}
}

// AddDevice adds a device to the DeviceManager device slice.
func (manager *deviceManager) AddDevice(device *Device) {
	manager.devices = append(manager.devices, device)
}

// AddDevices adds devices to the DeviceManager device slice.
func (manager *deviceManager) AddDevices(devices ...*Device) {
	manager.devices = append(manager.devices, devices...)
}

// RegisterDeviceSetupActions registers actions with the device manager which will be
// executed on plugin startup, prior to device loading but before plugin run. These
// actions are used for device-specific setup.
//
// fixme: no more kind, need to fix the below.
//
// The filter parameter should be the filter to apply to devices. Currently
// filtering is supported for device kind and type. Filter strings are specified in
// the format "key=value,key=value". The filter
//     "kind=temperature,kind=ABC123"
// would only match devices whose kind was temperature or ABC123.
func (manager *deviceManager) RegisterDeviceSetupActions(filter string, actions...*DeviceAction) {
	if _, exists := manager.setupActions[filter]; exists {
		manager.setupActions[filter] = append(manager.setupActions[filter], actions...)
	} else {
		manager.setupActions[filter] = actions
	}
}

// registerActions registers preRun (setup) and postRun (teardown) actions
// for the DeviceManager.
func (manager *deviceManager) registerActions(plugin *Plugin) {
	// Register pre-run actions.
	plugin.RegisterPreRunActions(
		&PluginAction{
			Name: "Load Device Configuration",
			Action: func(_ *Plugin) error { return manager.loadConfig() },
		},
		&PluginAction{
			Name: "Generate Devices From Configuration",
			Action: func(_ *Plugin) error { return manager.createDevices() },
		},
		&PluginAction{
			Name: "Run Device Setup Actions",
			Action: func(p *Plugin) error { return manager.execDeviceSetupActions(p) },
		},
	)
}

func (manager *deviceManager) createDevices() error {
	if manager.config == nil {
		// fixme: custom error?
		return fmt.Errorf("device manager has no config")
	}

	var failedLoad bool

	for _, proto := range manager.config.Devices {
		for _, instance := range proto.Instances {

			// Create the device.
			device, err := NewDeviceFromConfig(proto, instance)
			if err != nil {
				// todo: log
				failedLoad = true
				continue
			}
			// Add it to the manager.
			manager.AddDevice(device)
		}
	}

	if failedLoad {
		// fixme
		return fmt.Errorf("failed to load devices from config")
	}
	return nil
}

func (manager *deviceManager) loadConfig() error {
	// Setup the config loader for the device manager.
	loader := cfg.NewYamlLoader("device")
	loader.EnvOverride = DeviceEnvOverride
	loader.AddSearchPaths(
		"./config/device", // Local device config directory
		"/etc/synse/plugin/config/device", // Default device config directory
	)

	// Load the device configurations.
	if err := loader.Load(); err != nil {
		return err
	}

	return loader.Scan(manager.config)
}


func (manager *deviceManager) execDeviceSetupActions(plugin *Plugin) error {
	if len(manager.setupActions) == 0 {
		return nil
	}

	var multiErr = errors.NewMultiError("Device Setup Actions")

	log.WithFields(log.Fields{
		"actions": len(manager.setupActions),
	}).Info("[device manager] executing device setup actions")

	for filter, actions := range manager.setupActions {
		// todo: this will be updated to use devices from the device manager, not
		//  the global context thing.
		devices, err := filterDevices(filter)
		if err != nil {
			log.WithField("filter", filter).Error(
				"[device manager] failed to filter device for setup actions",
			)
			multiErr.Add(err)
			continue
		}

		log.WithFields(log.Fields{
			"matches": len(devices),
			"filter": filter,
		}).Debug("[device manager] applied filter to devices")
		for _, action := range actions {
			log.WithFields(log.Fields{
				"action": action.Name,
			}).Debug("[device manager] running device setup action")
			for _, device := range devices {
				if err := action.Action(plugin, device); err != nil {
					log.WithFields(log.Fields{
						"action": action.Name,
						"device": device.id,
					}).Error("[device manager] failed to run setup action for device")
					multiErr.Add(err)
					continue
				}
			}
		}
	}
	return multiErr.Err()
}