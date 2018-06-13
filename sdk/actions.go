package sdk

import (
	logger "github.com/Sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
)

type pluginAction func(p *Plugin) error
type deviceAction func(p *Plugin, d *Device) error

// execPreRun executes the pre-run actions for the plugin.
func execPreRun(plugin *Plugin) *errors.MultiError {
	var multiErr = errors.NewMultiError("pre-run actions")

	if len(ctx.preRunActions) > 0 {
		logger.Debug("[sdk] Executing pre-run actions:")
		for _, action := range ctx.preRunActions {
			logger.Debugf(" * %v", action)
			err := action(plugin)
			if err != nil {
				logger.Errorf("[sdk] Failed pre-run action %v: %v", action, err)
				multiErr.Add(err)
			}
		}
	}
	return multiErr
}

// execPostRun executes the post-run actions for the plugin.
func execPostRun(plugin *Plugin) *errors.MultiError {
	var multiErr = errors.NewMultiError("post-run actions")

	if len(ctx.postRunActions) > 0 {
		logger.Debug("[sdk] Executing post-run actions:")
		for _, action := range ctx.postRunActions {
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

	if len(ctx.deviceSetupActions) > 0 {
		logger.Debug("[sdk] Executing device setup actions:")
		for filter, acts := range ctx.deviceSetupActions {
			devices, err := filterDevices(filter)
			if err != nil {
				logger.Errorf("[sdk] Failed to filter devices for setup actions: %v", err)
				multiErr.Add(err)
				continue
			}
			logger.Debugf("* %v (%v devices match filter %v)", acts, len(devices), filter)
			for _, d := range devices {
				for _, action := range acts {
					err := action(plugin, d)
					if err != nil {
						logger.Errorf("[sdk] Failed device setup action %v: %v", action, err)
						multiErr.Add(err)
						continue
					}
				}
			}
		}
	}
	return multiErr
}
