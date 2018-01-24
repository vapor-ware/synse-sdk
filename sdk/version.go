package sdk

// SDKVersion specifies the version of the Synse Plugin SDK.
const SDKVersion = "0.3.2"

// VersionInfo contains the versioning information for a Plugin.
type VersionInfo struct {
	BuildDate     string
	GitCommit     string
	GitTag        string
	GoVersion     string
	VersionString string
}

// Merge merges the values in the given VersionInfo into this one.
// This is useful for merging a VersionConfig with the VersionConfig
// holding the default "empty" values.
func (v *VersionInfo) Merge(info *VersionInfo) {
	if info.BuildDate != "" {
		v.BuildDate = info.BuildDate
	}
	if info.GitCommit != "" {
		v.GitCommit = info.GitCommit
	}
	if info.GitTag != "" {
		v.GitTag = info.GitTag
	}
	if info.GoVersion != "" {
		v.GoVersion = info.GoVersion
	}
	if info.VersionString != "" {
		v.VersionString = info.VersionString
	}
}

// emptyVersionInfo gets a new VersionInfo instance filled with its
// default empty values.
func emptyVersionInfo() *VersionInfo {
	return &VersionInfo{
		BuildDate:     "-",
		GitCommit:     "-",
		GitTag:        "-",
		GoVersion:     "-",
		VersionString: "-",
	}
}
