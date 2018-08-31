package sdk

import (
	log "github.com/Sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
)

type pluginAction func(p *Plugin) error
type deviceAction func(p *Plugin, d *Device) error

// execPreRun executes the pre-run actions for the plugin.
func execPreRun(plugin *Plugin) *errors.MultiError {
	var multiErr = errors.NewMultiError("pre-run actions")

	log.Debugf("[sdk] executing %d pre-run action(s)", len(ctx.preRunActions))
	if len(ctx.preRunActions) > 0 {
		for _, action := range ctx.preRunActions {
			log.Debugf(" * %v", action)
			err := action(plugin)
			if err != nil {
				log.Errorf("[sdk] failed pre-run action %v: %v", action, err)
				multiErr.Add(err)
			}
		}
	}
	return multiErr
}

// execPostRun executes the post-run actions for the plugin.
func execPostRun(plugin *Plugin) *errors.MultiError {
	var multiErr = errors.NewMultiError("post-run actions")

	log.Debugf("[sdk] executing %d post-run action(s)", len(ctx.postRunActions))
	if len(ctx.postRunActions) > 0 {
		for _, action := range ctx.postRunActions {
			log.Debugf(" * %v", action)
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

	log.Debugf("[sdk] executing %d device setup action(s)", len(ctx.deviceSetupActions))
	if len(ctx.deviceSetupActions) > 0 {
		for filter, acts := range ctx.deviceSetupActions {
			devices, err := filterDevices(filter)
			if err != nil {
				log.Errorf("[sdk] failed to filter devices for setup actions: %v", err)
				multiErr.Add(err)
				continue
			}
			log.Debugf("* %v (%v devices match filter %v)", acts, len(devices), filter)
			for _, d := range devices {
				for _, action := range acts {
					err := action(plugin, d)
					if err != nil {
						log.Errorf("[sdk] failed device setup action %v: %v", action, err)
						multiErr.Add(err)
						continue
					}
				}
			}
		}
	}
	return multiErr
}
