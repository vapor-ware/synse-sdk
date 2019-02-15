package sdk

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/vapor-ware/synse-server-grpc/go"

	log "github.com/Sirupsen/logrus"
)

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
  Tag:         {{.Tag}}
  Name:        {{.Name}}
  Maintainer:  {{.Maintainer}}
  VCS:         {{.VCS}}
  Description: {{.Description}}`

	t := template.Must(template.New("metadata").Parse(out))
	_ = t.Execute(&writer, version) // nolint

	return writer.String()
}
