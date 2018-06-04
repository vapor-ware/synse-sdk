package sdk

import (
	"runtime"

	"bytes"
	"text/template"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
)

// SDKVersion specifies the version of the Synse Plugin SDK.
const SDKVersion = "1.0.0"

// version is a reference to a BinVersion that can be used to get
// the version info for a plugin.
var Version *BinVersion

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
	Version = &BinVersion{
		Arch:          runtime.GOARCH,
		OS:            runtime.GOOS,
		SDKVersion:    SDKVersion,
		BuildDate:     setField(BuildDate),
		GitCommit:     setField(GitCommit),
		GitTag:        setField(GitTag),
		GoVersion:     setField(GoVersion),
		PluginVersion: setField(PluginVersion),
	}
}


// BinVersion describes the version of the binary for a plugin.
//
// This should be populated via build-time args passed in for
// the corresponding variables.
type BinVersion struct {
	Arch          string
	BuildDate     string
	GitCommit     string
	GitTag        string
	GoVersion     string
	OS            string
	PluginVersion string
	SDKVersion    string
}

// Format returns a formatted string with all of the BinVersion info.
func (version *BinVersion) Format() string {
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

// Log logs out the BinVersion at info level.
func (version *BinVersion) Log() {
	logger.Info("Version Info:")
	logger.Infof("  Plugin Version: %s", version.PluginVersion)
	logger.Infof("  SDK Version:    %s", version.SDKVersion)
	logger.Infof("  Git Commit:     %s", version.GitCommit)
	logger.Infof("  Git Tag:        %s", version.GitTag)
	logger.Infof("  Build Date:     %s", version.BuildDate)
	logger.Infof("  Go Version:     %s", version.GoVersion)
	logger.Infof("  OS/Arch:        %s/%s", version.OS, version.Arch)
}

// setField is a helper function that checks whether a field is set.
// If the field is set, that field is returned, otherwise "-" is returned.
func setField(field string) string {
	if field == "" {
		return "-"
	}
	return field
}
