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
	"bytes"
	"fmt"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
	synse "github.com/vapor-ware/synse-server-grpc/go"
)

// metadata is a plugin-global reference to the plugin metadata.
var metadata PluginMetadata

// SetPluginInfo sets the meta-information for a Plugin. This should
// be this first step in creating a new plugin.
func SetPluginInfo(name, maintainer, desc, vcs string) {
	metadata = PluginMetadata{
		Name:        name,
		Maintainer:  maintainer,
		Description: desc,
		VCS:         vcs,
	}
}

// PluginMeta is the metadata associated with a Plugin.
type PluginMetadata struct {
	Name        string
	Maintainer  string
	Description string
	VCS         string
}

// Tag creates the tag used in the plugin meta information.
func (info *PluginMetadata) Tag() string {
	tag := fmt.Sprintf("%s/%s", info.Maintainer, info.Name)
	tag = strings.ToLower(tag)
	tag = strings.Replace(tag, "-", "_", -1)
	tag = strings.Replace(tag, " ", "-", -1)
	return tag
}

// log logs out the plugin metadata at INFO level.
func (info *PluginMetadata) log() {
	log.Info("Plugin Info:")
	log.Infof("  Tag:         %s", info.Tag())
	log.Infof("  Name:        %s", info.Name)
	log.Infof("  Maintainer:  %s", info.Maintainer)
	log.Infof("  VCS:         %s", info.VCS)
	log.Infof("  Description: %s", info.Description)
}

// encode converts the metadata struct to its corresponding Synse gRPC message.
func (info *PluginMetadata) encode() *synse.V3Metadata {
	return &synse.V3Metadata{
		Name:        info.Name,
		Maintainer:  info.Maintainer,
		Tag:         info.Tag(),
		Description: info.Description,
		Vcs:         info.VCS,
	}
}

// format returns a formatted string with the plugin metadata.
func (info *PluginMetadata) format() string {
	var writer bytes.Buffer

	out := `Plugin Info:
  Name:        {{.Name}}
  Maintainer:  {{.Maintainer}}
  VCS:         {{.VCS}}
  Description: {{.Description}}`

	t := template.Must(template.New("metadata").Parse(out))
	_ = t.Execute(&writer, info) // nolint

	return writer.String()
}
