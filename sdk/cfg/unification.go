package cfg

import (
	"fmt"

	"github.com/vapor-ware/synse-sdk/sdk/errors"
)

// UnifyDeviceConfigs will take a slice of ConfigContext which represents
// DeviceConfigs and unify them into a single ConfigContext for a DeviceConfig.
//
// If any of the ConfigContexts given as a parameter do not represent a
// DeviceConfig, an error is returned.
func UnifyDeviceConfigs(ctxs []*ConfigContext) (*ConfigContext, error) {

	// FIXME (etd): figure out how to either:
	//  i. merge the source info into the ConfigContext
	// ii. map each component to its original context so we know exactly where
	//     a specific field/config component originated from.

	// If there are no contexts, we can't unify.
	if len(ctxs) == 0 {
		return nil, fmt.Errorf("no ConfigContexts specified for unification")
	}

	var context *ConfigContext

	for _, ctx := range ctxs {
		if !ctx.IsDeviceConfig() {
			return nil, fmt.Errorf("config context does not represent a device config")
		}
		if context == nil {
			context = ctx
		} else {
			base := context.Config.(*DeviceConfig)
			source := ctx.Config.(*DeviceConfig)

			// Merge DeviceConfig.Locations
			base.Locations = append(base.Locations, source.Locations...)

			// Merge DeviceConfig.Devices
			base.Devices = append(base.Devices, source.Devices...)
		}
	}

	return context, nil
}
