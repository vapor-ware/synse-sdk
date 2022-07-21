// Synse SDK
// Copyright (c) 2017-2020 Vapor IO
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
	"runtime"
	"text/template"

	log "github.com/sirupsen/logrus"
	synse "github.com/vapor-ware/synse-server-grpc/go"
)

// Version specifies the version of the Synse Plugin SDK.
const Version = "v2.1.2"

// version is a global reference to the pluginVersion which specifies the
// version information for a Plugin. This is initialized on init and
// populated with build-time arguments.
var version *pluginVersion

var (
	// BuildDate is the timestamp for when the build happened.
	BuildDate string

	// GitCommit is the commit hash at which the plugin was built.
	GitCommit string

	// GitTag is the git tag at which the plugin was built.
	GitTag string

	// GoVersion is is the version of Go used to build the plugin.
	GoVersion string

	// PluginVersion is the canonical version string for the plugin.
	PluginVersion string
)

func init() {
	version = &pluginVersion{
		Arch:          runtime.GOARCH,
		OS:            runtime.GOOS,
		SDKVersion:    Version,
		BuildDate:     setField(BuildDate),
		GitCommit:     setField(GitCommit),
		GitTag:        setField(GitTag),
		GoVersion:     setField(GoVersion),
		PluginVersion: setField(PluginVersion),
	}
}

// pluginVersion describes the version of a Synse plugin.
type pluginVersion struct {
	Arch          string
	BuildDate     string
	GitCommit     string
	GitTag        string
	GoVersion     string
	OS            string
	PluginVersion string
	SDKVersion    string
}

// encode converts the pluginVersion to its corresponding Synse gRPC message.
func (version *pluginVersion) encode() *synse.V3Version {
	return &synse.V3Version{
		PluginVersion: version.PluginVersion,
		SdkVersion:    version.SDKVersion,
		BuildDate:     version.BuildDate,
		GitCommit:     version.GitCommit,
		GitTag:        version.GitTag,
		Arch:          version.Arch,
		Os:            version.OS,
	}
}

// format returns a formatted string with all of the version info.
func (version *pluginVersion) format() string {
	var info bytes.Buffer

	out := `Version Info:
  Plugin Version: {{.PluginVersion}}
  SDK Version:    {{.SDKVersion}}
  Git Commit:     {{.GitCommit}}
  Git Tag:        {{.GitTag}}
  Build Date:     {{.BuildDate}}
  Go Version:     {{.GoVersion}}
  OS/Arch:        {{.OS}}/{{.Arch}}`

	t := template.Must(template.New("version").Parse(out))
	_ = t.Execute(&info, version) // nolint

	return info.String()
}

// Log logs out the version information at info level.
func (version *pluginVersion) Log() {
	log.Info("Version Info:")
	log.Infof("  Plugin Version: %s", version.PluginVersion)
	log.Infof("  SDK Version:    %s", version.SDKVersion)
	log.Infof("  Git Commit:     %s", version.GitCommit)
	log.Infof("  Git Tag:        %s", version.GitTag)
	log.Infof("  Build Date:     %s", version.BuildDate)
	log.Infof("  Go Version:     %s", version.GoVersion)
	log.Infof("  OS/Arch:        %s/%s", version.OS, version.Arch)
}

// setField is a helper function that checks whether a version field is set.
// If the field is set, that field is returned, otherwise "-" is returned.
func setField(field string) string {
	if field == "" {
		return "-"
	}
	return field
}
