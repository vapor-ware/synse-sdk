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
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/policy"
)

const (
	// DeviceEnvOverride defines the environment variable that can be used to
	// set an override config location for device configuration files.
	DeviceEnvOverride = "PLUGIN_DEVICE_CONFIG"
)

// todo: figure out where dynamic device registration fits in here.

// DeviceAction defines an action that can be run before the main Plugin run
// logic. This is generally used for doing device-specific setup actions.
type DeviceAction struct {
	// Name is the name of the action. This is used to identify the action.
	Name string

	// Filter is the device filter that scopes which devices this action
	// should apply to. This filter is run on the entire set of registered
	// devices and is additive (e.g. a device does not need to match all
	// filters to be included, it just needs to match one).
	//
	// The filter provided should be a map, where the key is the field to filter
	// on and the value are the allowable values for that field. The currently
	// supported filters include:
	//  * "type" : the device type
	Filter map[string][]string

	// The action to execute for the device.
	Action func(p *Plugin, d *Device) error
}

// deviceManager loads and manages a Plugin's devices.
type deviceManager struct {
	config         *config.Devices
	id             *pluginID
	pluginHandlers *PluginHandlers
	policies       *policy.Policies

	tagCache     *TagCache
	setupActions []*DeviceAction
	devices      map[string]*Device
	handlers     map[string]*DeviceHandler
}

// newDeviceManager creates a new DeviceManager.
// fixme (etd): this constructor will be cleaned up in the future. instead of passing everything in
//  one at a time, we can pass in some sort of context which has everything it needs...
func newDeviceManager(id *pluginID, handlers *PluginHandlers, policies *policy.Policies) *deviceManager {
	return &deviceManager{
		config:         new(config.Devices),
		id:             id,
		pluginHandlers: handlers,
		policies:       policies,
		tagCache:       NewTagCache(),
		devices:        make(map[string]*Device),
		handlers:       make(map[string]*DeviceHandler),
	}
}

// init is the initialization function for the deviceManager. This ensures that
// the device config is loaded and that the config is parsed into the appropriate
// Device models.
func (manager *deviceManager) init() error {
	log.Info("[device manager] initializing")

	if err := manager.loadConfig(); err != nil {
		return err
	}

	if err := manager.createDevices(); err != nil {
		return err
	}

	return nil
}

// Start starts the deviceManager.
//
// Unlike other components, there is no long-running action which will be kicked
// off here. This is just where device setup actions are executed. This should be
// done here rather than in init.
func (manager *deviceManager) Start(plugin *Plugin) error {
	log.Info("[device manager] starting")
	return manager.execDeviceSetupActions(plugin)
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

// GetDevicesByTagNamespace gets all the devices in the specified tag namespace(s).
func (manager *deviceManager) GetDevicesByTagNamespace(namespace ...string) []*Device {
	return manager.tagCache.GetDevicesFromNamespace(namespace...)
}

// GetAllDevices gets all devices that are registered with the deviceManager.
func (manager *deviceManager) GetAllDevices() []*Device {
	devices := make([]*Device, 0, len(manager.devices))
	for _, device := range manager.devices {
		devices = append(devices, device)
	}
	return devices
}

// IsDeviceReadable checks whether a given device is readable.
func (manager *deviceManager) IsDeviceReadable(id string) bool {
	device := manager.GetDevice(id)
	return device.IsReadable()
}

// IsDeviceWritable checks whether a given device is writable.
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

	// TODO: verify device Data with the custom verification fn

	// If the device ID has not already been set, generate it and set
	// it before adding it to the deviceManager.
	if device.id == "" {
		// todo: see about cleaning this up/making it its own fn so it can be reused.
		component := manager.pluginHandlers.DeviceIdentifier(device.Data)
		name := strings.Join([]string{
			device.Type,
			device.Handler,
			component,
		}, ".")
		device.idName = name

		deviceID := manager.id.NewNamespacedID(name)
		device.id = deviceID
	}

	// Check if the Device ID collides with an existing device.
	if _, exists := manager.devices[device.id]; exists {
		// fixme
		return fmt.Errorf("device id exists")
	}

	// Update the device with the SDK auto-generated tags.
	device.Tags = append(device.Tags, newIDTag(device.id), newTypeTag(device.Type))

	// Add the device to the manager.
	manager.devices[device.id] = device

	// Update the tag cache for the device.
	for _, t := range device.Tags {
		manager.tagCache.Add(t, device)
	}

	log.WithFields(log.Fields{
		"id":   device.id,
		"type": device.Type,
	}).Info("[device manager] added new device")

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
// A DeviceAction should specify a filter which is used to target the devices which
// the action should apply to. If a DeviceAction does not have a filter, it will
// not be accepted by the deviceManager.
func (manager *deviceManager) AddDeviceSetupActions(actions ...*DeviceAction) error {
	for _, action := range actions {
		if len(action.Filter) == 0 {
			log.WithFields(log.Fields{
				"action": action.Name,
			}).Error("[device manager] no filter set for device setup action")
		}
		manager.setupActions = append(manager.setupActions, action)
	}
	return nil
}

// FilterDevices applies a filter to the compete set of registered devices and returns
// the set of devices which match the filter.
//
// The filter provided should be a map, where the key is the field to filter on and the
// value are the allowable values for that field. This filtering is additive, e.g.
// type=temperature and type=led will return all temperature and led devices.
func (manager *deviceManager) FilterDevices(filter map[string][]string) ([]*Device, error) {
	var filteredSet []*Device
	var checks []func(d *Device) bool

	for k, v := range filter {
		var check func(d *Device) bool

		// todo: support more filters...
		switch k {
		case "type":
			check = func(d *Device) bool {
				for _, val := range v {
					if d.Type == val || val == "*" {
						return true
					}
				}
				return false
			}
		default:
			// fixme: better errors
			return nil, fmt.Errorf("unsupported filter key")
		}

		checks = append(checks, check)
	}

	for _, device := range manager.devices {
		for _, check := range checks {
			if check(device) {
				filteredSet = append(filteredSet, device)
			}
		}
	}
	return filteredSet, nil
}

// createDevices takes the manager configuration and generates all corresponding
// Device instances from it.
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
				log.WithField("error", err).Error("[device manager] failed to create device from config")
				failedLoad = true
				continue
			}
			// Add it to the manager.
			if err := manager.AddDevice(device); err != nil {
				// todo: log
				log.WithField("error", err).Error("[device manager] failed to add device to manager")
				failedLoad = true
			}
		}
	}

	if failedLoad {
		// fixme
		log.Errorf("[device manager] failed to create devices from config")
		return fmt.Errorf("failed to load devices from config")
	}

	log.WithField("devices", len(manager.devices)).Info("[device manager] created devices")
	return nil
}

// loadConfig is a helper function used to load device configurations into the
// deviceManager.
func (manager *deviceManager) loadConfig() error {
	// Setup the config loader for the device manager.
	loader := config.NewYamlLoader("device")
	loader.EnvOverride = DeviceEnvOverride
	loader.AddSearchPaths(
		"./config/device",                 // Local device config directory
		"/etc/synse/plugin/config/device", // Default device config directory
	)

	// Load the device configurations.
	if err := loader.Load(manager.policies.DeviceConfig); err != nil {
		return err
	}

	return loader.Scan(manager.config)
}

// execDeviceStartupActions runs all the device startup actions registered with
// the manager. This should be done before any reads/write occur (e.g. before
// the scheduler is started).
func (manager *deviceManager) execDeviceSetupActions(plugin *Plugin) error {
	if len(manager.setupActions) == 0 {
		return nil
	}

	var multiErr = errors.NewMultiError("Device Setup Actions")

	log.WithFields(log.Fields{
		"actions": len(manager.setupActions),
	}).Info("[device manager] executing device setup actions")

	for _, action := range manager.setupActions {
		devices, err := manager.FilterDevices(action.Filter)
		if err != nil {
			log.WithField("filter", action.Filter).Error(
				"[device manager] failed to filter device for setup actions",
			)
			multiErr.Add(err)
			continue
		}

		log.WithFields(log.Fields{
			"action":  action.Name,
			"matches": len(devices),
			"filter":  action.Filter,
		}).Debug("[device manager] applied filter to devices")

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
	return multiErr.Err()
}
