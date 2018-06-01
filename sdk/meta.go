package sdk

import (
	"github.com/vapor-ware/synse-sdk/sdk/logger"
)

// metainfo is the global variable that tracks plugin meta-information.
var metainfo meta

// meta is a struct that holds the meta-information for a plugin.
type meta struct {
	Name        string
	Maintainer  string
	Description string
	VCS         string
}

// log logs out the plugin meta-info at INFO level.
func (m *meta) log() {
	logger.Info("Plugin Info:")
	logger.Infof("  Name:        %s", m.Name)
	logger.Infof("  Maintainer:  %s", m.Maintainer)
	logger.Infof("  Description: %s", m.Description)
	logger.Infof("  VCS:         %s", m.VCS)
}

// SetPluginMeta sets the meta-information for a plugin.
func SetPluginMeta(name, maintainer, desc, vcs string) {
	metainfo = meta{
		Name:        name,
		Maintainer:  maintainer,
		Description: desc,
		VCS:         vcs,
	}
}
