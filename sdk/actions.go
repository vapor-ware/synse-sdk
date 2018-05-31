package sdk

import (
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
)

type pluginAction func(p *Plugin) error
type deviceAction func(p *Plugin, d *Device) error

func execPreRun(plugin *Plugin) *errors.MultiError {
	var multiErr = errors.NewMultiError("Pre Run Actions")

	if len(plugin.preRunActions) > 0 {
		logger.Debug("Executing Pre Run Actions:")
		for _, action := range plugin.preRunActions {
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

func execDeviceSetup(plugin *Plugin) *errors.MultiError {
	var multiErr = errors.NewMultiError("Device Setup Actions")

	if len(plugin.deviceSetupActions) > 0 {
		logger.Debug("Executing Device Setup Actions:")
		for filter, acts := range plugin.deviceSetupActions {
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
