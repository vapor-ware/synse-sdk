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

// TODO (etd) -- the below are really unification mixed with verification checks. Move the
// verification stuff out of here.

func mergeLocations(base, source *DeviceConfig) *errors.MultiError {
	var multiErr = errors.NewMultiError("Device Config Location Merge")
	var toMerge []*Location
	var conflict bool

	for _, sourceLocation := range source.Locations {
		for _, baseLocation := range base.Locations {
			// FIXME: this should check if the location data is the same. if all location info
			// is the same, then do not error. if the names are the same but the info is different,
			// then error.
			if sourceLocation.Name == baseLocation.Name {
				conflict = true
				multiErr.Add(fmt.Errorf("unify conflict: multiple Locations specified with the same name: %s", sourceLocation.Name))
			}
		}
		// If there is no conflict, we can update the toMerge list.
		// If there is a conflict, we will not be merging, so there is
		// no need to track what should be merged.
		if !conflict {
			toMerge = append(toMerge, sourceLocation)
		}
	}

	// If there are no conflicts, update.
	if !conflict {
		base.Locations = append(base.Locations, toMerge...)
	}
	return multiErr
}

func mergeKinds(base, source *DeviceConfig) *errors.MultiError {
	var multiErr = errors.NewMultiError("Device Config DeviceKind Merge")
	var toMerge []*DeviceKind
	var conflict bool

	for _, sourceKind := range source.Devices {
		for _, baseKind := range base.Devices {
			// FIXME: should unification care about this? Or should this be a check for verification?
			if sourceKind.Name == baseKind.Name {
				conflict = true
				multiErr.Add(fmt.Errorf("unify conflict: multiple DeviceKind share the same name: %s", sourceKind.Name))
			}
		}

		// If there is no conflict, we can update the toMerge list.
		if !conflict {
			toMerge = append(toMerge, sourceKind)
		}
	}

	if !conflict {
		base.Devices = append(base.Devices, toMerge...)
	}
	return multiErr
}
