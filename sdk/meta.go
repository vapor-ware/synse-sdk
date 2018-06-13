package sdk

import (
	log "github.com/Sirupsen/logrus"
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
	log.Info("Plugin Info:")
	log.Infof("  Name:        %s", m.Name)
	log.Infof("  Maintainer:  %s", m.Maintainer)
	log.Infof("  Description: %s", m.Description)
	log.Infof("  VCS:         %s", m.VCS)
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
