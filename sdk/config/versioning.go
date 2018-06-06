package config

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	tagAddedIn      = "addedIn"
	tagDeprecatedIn = "deprecatedIn"
	tagRemovedIn    = "removedIn"
)

// Version is a representation of a configuration scheme version
// that can be compared to other SchemeVersions.
type Version struct {
	Major int
	Minor int
}

// NewVersion creates a new instance of a Version.
func NewVersion(versionString string) (*Version, error) {
	var min, maj int
	var err error

	if versionString == "" {
		return nil, fmt.Errorf("no version info found")
	}

	s := strings.Split(versionString, ".")
	switch len(s) {
	case 1:
		maj, err = strconv.Atoi(s[0])
		if err != nil {
			return nil, err
		}
		min = 0
	case 2:
		maj, err = strconv.Atoi(s[0])
		if err != nil {
			return nil, err
		}
		min, err = strconv.Atoi(s[1])
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("too many version components - should only have MAJOR[.MINOR]")
	}

	return &Version{
		Major: maj,
		Minor: min,
	}, nil
}

// String returns a string representation of the scheme version.
func (version *Version) String() string {
	return fmt.Sprintf("%d.%d", version.Major, version.Minor)
}

// IsLessThan returns true if the Version is less than the Version
// provided as a parameter.
func (version *Version) IsLessThan(other *Version) bool {
	if version.Major < other.Major {
		return true
	}
	if version.Major == other.Major && version.Minor < other.Minor {
		return true
	}
	return false
}

// IsGreaterOrEqualTo returns true if the Version is greater than or equal to
// the Version provided as a parameter.
func (version *Version) IsGreaterOrEqualTo(other *Version) bool {
	if version.Major > other.Major {
		return true
	}
	if version.Major == other.Major && version.Minor >= other.Minor {
		return true
	}
	return false
}

// IsEqual returns true if the Version is equal to the Version provided
// as a parameter.
func (version *Version) IsEqual(other *Version) bool {
	return version.Major == other.Major && version.Minor == other.Minor
}

// SchemeVersion is a struct that is used to extract the configuration
// scheme version from any config file.
type SchemeVersion struct {
	// Version is the config version scheme specified in the config file.
	Version string `yaml:"version,omitempty" addedIn:"1.0"`

	// scheme is the Version that represents the SchemeVersion's Version.
	scheme *Version
}

// parse parses the Version field into a Version.
func (schemeVersion *SchemeVersion) parse() error {
	scheme, err := NewVersion(schemeVersion.Version)
	if err != nil {
		return err
	}
	schemeVersion.scheme = scheme
	return nil
}

// GetVersion gets the Version associated with the version specified
// in the configuration.
func (schemeVersion *SchemeVersion) GetVersion() (*Version, error) {
	if schemeVersion.scheme == nil {
		err := schemeVersion.parse()
		if err != nil {
			return nil, err
		}
	}
	return schemeVersion.scheme, nil
}
