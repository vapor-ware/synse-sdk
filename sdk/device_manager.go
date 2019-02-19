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

	log "github.com/Sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
)

const (
	// DeviceEnvOverride defines the environment variable that can be used to
	// set an override config location for device configuration files.
	DeviceEnvOverride = "PLUGIN_DEVICE_CONFIG"
)

// todo: figure out where dynamic device registration fits in here.

// todo: exec device startup actions.. probably not in init (just for startup
//  ordering), but could be done in a start() fn.

// DeviceAction defines an action that can be run before the main Plugin run
// logic. This is generally used for doing device-specific setup actions.
type DeviceAction struct {
	Name   string
	Action func(p *Plugin, d *Device) error
}

// deviceManager loads and manages a Plugin's devices.
type deviceManager struct {
	config *config.Devices

	tagCache     *TagCache
	devices      map[string]*Device
	handlers     map[string]*DeviceHandler
	setupActions map[string][]*DeviceAction
}

// newDeviceManager creates a new DeviceManager.
func newDeviceManager() *deviceManager {
	return &deviceManager{
		tagCache:     NewTagCache(),
		devices:      make(map[string]*Device),
		handlers:     make(map[string]*DeviceHandler),
		setupActions: make(map[string][]*DeviceAction),
	}
}

func (manager *deviceManager) init() error {
	if err := manager.loadConfig(); err != nil {
		return err
	}

	if err := manager.createDevices(); err != nil {
		return err
	}

	return nil
}

// GetDevice gets a device from the manager by ID.
func (manager *deviceManager) GetDevice(id string) *Device {
	device, exists := manager.devices[id]
	if !exists {
		log.WithFields(log.Fields{
			"id": id,
		}).Debug("[device manager] device does not exist")
	}
	return device
}

// GetDevices gets all devices which match the given set of tags.
func (manager *deviceManager) GetDevices(tags ...*Tag) []*Device {
	return manager.tagCache.GetDevicesFromTags(tags...)
}

func (manager *deviceManager) IsDeviceReadable(id string) bool {
	device := manager.GetDevice(id)
	return device.IsReadable()
}

func (manager *deviceManager) IsDeviceWritable(id string) bool {
	device := manager.GetDevice(id)
	return device.IsWritable()
}

// HasReadHandlers checks whether any of the DeviceHandlers registered with
// the deviceManager implement a read function.
func (manager *deviceManager) HasReadHandlers() bool {
	for _, h := range manager.handlers {
		if h.Read != nil || h.BulkRead != nil {
			return true
		}
	}
	return false
}

// HasWriteHandlers checks whether any of the DeviceHandlers registered with
// the deviceManager implement a write function.
func (manager *deviceManager) HasWriteHandlers() bool {
	for _, h := range manager.handlers {
		if h.Write != nil {
			return true
		}
	}
	return false
}

// HasListenerHandlers checks whether any of the DeviceHandlers registered with
// the deviceManager implement a listener function.
func (manager *deviceManager) HasListenerHandlers() bool {
	for _, h := range manager.handlers {
		if h.Listen != nil {
			return true
		}
	}
	return false
}

// AddDevice adds a device to the DeviceManager and makes sure that it
// has a reference to its DeviceHandler.
//
// All devices should be added to the DeviceManager with this function.
//
// If the Device specifies a handler that does not exist, this will
// result in an error.
func (manager *deviceManager) AddDevice(device *Device) error {
	if device == nil {
		return fmt.Errorf("can not add nil device to device manager")
	}
	if device.Handler == "" {
		return fmt.Errorf("device does not specify a Handler, can not add to device manager")
	}

	// If the device does not have a handler set, look up the handler and
	// update the Device instance with a reference to that handler.
	if device.handler == nil {
		handler, err := manager.GetHandler(device.Handler)
		if err != nil {
			return err
		}
		device.handler = handler
	}

	// Check if the Device ID collides with an existing device.
	if _, exists := manager.devices[device.ID()]; exists {
		// fixme
		return fmt.Errorf("device id exists")
	}

	// Add the device to the manager.
	manager.devices[device.ID()] = device

	// Update the tag cache for the device.
	for _, t := range device.Tags {
		manager.tagCache.Add(t, device)
	}

	return nil
}

// AddHandlers adds DeviceHandlers to the DeviceManager.
func (manager *deviceManager) AddHandlers(handlers ...*DeviceHandler) error {
	for _, handler := range handlers {
		if _, exists := manager.handlers[handler.Name]; exists {
			return fmt.Errorf(
				"unable to register multiple handlers with duplicate names: %s",
				handler.Name,
			)
		}
		manager.handlers[handler.Name] = handler
	}
	return nil
}

// GetDevicesForHandler gets all of the Devices which are configured to use the
// DeviceHandler with the given name.
func (manager *deviceManager) GetDevicesForHandler(handler string) []*Device {
	var devices []*Device
	for _, device := range manager.devices {
		if device.Handler == handler {
			devices = append(devices, device)
		}
	}
	return devices
}

// GetHandler gets a DeviceHandler by name. If the named DeviceHandler does not
// exist, an error is returned.
func (manager *deviceManager) GetHandler(name string) (*DeviceHandler, error) {
	handler, exists := manager.handlers[name]
	if !exists {
		return nil, fmt.Errorf("device handler '%s' does not exist", name)
	}
	return handler, nil
}

// AddDeviceSetupActions registers actions with the device manager which will be
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
func (manager *deviceManager) AddDeviceSetupActions(filter string, actions ...*DeviceAction) {
	if _, exists := manager.setupActions[filter]; exists {
		manager.setupActions[filter] = append(manager.setupActions[filter], actions...)
	} else {
		manager.setupActions[filter] = actions
	}
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
			if err := manager.AddDevice(device); err != nil {
				// todo: log
				failedLoad = true
			}
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
	loader := config.NewYamlLoader("device")
	loader.EnvOverride = DeviceEnvOverride
	loader.AddSearchPaths(
		"./config/device",                 // Local device config directory
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
			"filter":  filter,
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
