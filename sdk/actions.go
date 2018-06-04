package sdk

import (
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
)

var (
	// preRunActions holds all of the known plugin actions to run prior to starting
	// up the plugin server and data manager.
	preRunActions []pluginAction

	// postRunActions holds all of the known plugin actions to run after terminating
	// the plugin server and data manager.
	postRunActions []pluginAction

	// deviceSetupActions holds all of the known device device setup actions to run
	// prior to starting up the plugin server and data manager. The map key is the
	// filter used to apply the deviceAction value to a Device instance.
	deviceSetupActions map[string][]deviceAction
)

func init() {
	// Initialize the global variables so they are never nil.
	preRunActions = []pluginAction{}
	postRunActions = []pluginAction{}
	deviceSetupActions = map[string][]deviceAction{}
}

type pluginAction func(p *Plugin) error
type deviceAction func(p *Plugin, d *Device) error

// execPreRun executes the pre-run actions for the plugin.
func execPreRun(plugin *Plugin) *errors.MultiError {
	var multiErr = errors.NewMultiError("pre-run actions")

	if len(preRunActions) > 0 {
		logger.Debug("Executing pre-run actions:")
		for _, action := range preRunActions {
			logger.Debugf(" * %v", action)
			err := action(plugin)
			if err != nil {
				logger.Errorf("Failed pre-run action %v: %v", action, err)
				multiErr.Add(err)
			}
		}
	}
	return multiErr
}

// execPostRun executes the post-run actions for the plugin.
func execPostRun(plugin *Plugin) *errors.MultiError {
	var multiErr = errors.NewMultiError("post-run actions")

	if len(postRunActions) > 0 {
		logger.Debug("Executing post-run actions:")
		for _, action := range postRunActions {
			logger.Debug(" * %v", action)
			err := action(plugin)
			if err != nil {
				multiErr.Add(err)
			}
		}
	}
	return multiErr
}

// execDeviceSetup executes the device setup actions for the plugin.
func execDeviceSetup(plugin *Plugin) *errors.MultiError {
	var multiErr = errors.NewMultiError("device setup actions")

	if len(deviceSetupActions) > 0 {
		logger.Debug("Executing device setup actions:")
		for filter, acts := range deviceSetupActions {
			devices, err := filterDevices(filter)
			if err != nil {
				logger.Errorf("Failed to filter devices for setup actions: %v", err)
				multiErr.Add(err)
				continue
			}
			logger.Debugf("* %v (%v devices match filter %v)", acts, len(devices), filter)
			for _, d := range devices {
				for _, action := range acts {
					err := action(plugin, d)
					if err != nil {
						logger.Errorf("Failed device setup action %v: %v", action, err)
						multiErr.Add(err)
						continue
					}
				}
			}
		}
	}
	return multiErr
}
