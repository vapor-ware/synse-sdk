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
)

const (
	// DeviceEnvOverride defines the environment variable that can be used to
	// set an override config location for device configuration files.
	DeviceEnvOverride = "PLUGIN_DEVICE_CONFIG"
)

var DeviceManager *deviceManager


func GetDeviceByID() {}
func GetDeviceByAlias() {}
func GetDevices() {}
func GetDevicesForHandler() {}


// todo: figure out where dynamic device registration fits in here.


// DeviceManager loads and manages a Plugin's devices.
type deviceManager struct {
	config  *cfg.Devices

	devices []*Device
}

// NewDeviceManager creates a new DeviceManager.
func NewDeviceManager() (*deviceManager, error) {

	// Load the device configurations.
	conf := new(cfg.Devices)
	if err := loadDeviceConfigs(conf); err != nil {
		return nil, err
	}

	manager := deviceManager{
		config: conf,
	}
	return &manager, nil
}

// AddDevice adds a device to the DeviceManager device slice.
func (manager *deviceManager) AddDevice(device *Device) {
	manager.devices = append(manager.devices, device)
}

// AddDevices adds devices to the DeviceManager device slice.
func (manager *deviceManager) AddDevices(devices ...*Device) {
	manager.devices = append(manager.devices, devices...)
}

// registerActions registers preRun (setup) and postRun (teardown) actions
// for the DeviceManager.
func (manager *deviceManager) registerActions(plugin *Plugin) {
	// Register pre-run actions.
	plugin.RegisterPreRunActions(
		func(plugin *Plugin) error { return manager.loadDevices() },
	)
}

func (manager *deviceManager) loadDevices() error {
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

// loadDeviceConfigs loads the configuration for Plugin devices.
func loadDeviceConfigs(conf *cfg.Devices) error {
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

	return loader.Scan(conf)
}