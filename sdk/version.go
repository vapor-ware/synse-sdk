package sdk

import (
	"bytes"
	"runtime"
	"text/template"

	log "github.com/Sirupsen/logrus"
	"github.com/vapor-ware/synse-server-grpc/go"
)

// Version specifies the version of the Synse Plugin SDK.
const Version = "1.2.0"

// version is a reference to a binVersion that is used by the SDK to get
// the version info for a plugin.
var version *binVersion

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
	version = &binVersion{
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

// binVersion describes the version of the binary for a plugin.
//
// This should be populated via build-time args passed in for
// the corresponding variables.
type binVersion struct {
	Arch          string
	BuildDate     string
	GitCommit     string
	GitTag        string
	GoVersion     string
	OS            string
	PluginVersion string
	SDKVersion    string
}

// encode converts the binVersion to its corresponding Synse GRPC VersionInfo message.
func (version *binVersion) Encode() *synse.VersionInfo {
	return &synse.VersionInfo{
		PluginVersion: version.PluginVersion,
		SdkVersion:    version.SDKVersion,
		BuildDate:     version.BuildDate,
		GitCommit:     version.GitCommit,
		GitTag:        version.GitTag,
		Arch:          version.Arch,
		Os:            version.OS,
	}
}

// Format returns a formatted string with all of the binVersion info.
func (version *binVersion) Format() string {
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

// Log logs out the binVersion at info level.
func (version *binVersion) Log() {
	log.Info("Version Info:")
	log.Infof("  Plugin Version: %s", version.PluginVersion)
	log.Infof("  SDK Version:    %s", version.SDKVersion)
	log.Infof("  Git Commit:     %s", version.GitCommit)
	log.Infof("  Git Tag:        %s", version.GitTag)
	log.Infof("  Build Date:     %s", version.BuildDate)
	log.Infof("  Go Version:     %s", version.GoVersion)
	log.Infof("  OS/Arch:        %s/%s", version.OS, version.Arch)
}

// setField is a helper function that checks whether a field is set.
// If the field is set, that field is returned, otherwise "-" is returned.
func setField(field string) string {
	if field == "" {
		return "-"
	}
	return field
}
