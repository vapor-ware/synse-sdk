package cfg

import (
	"fmt"
	"strconv"
	"strings"
)

// Comparison is a type that represents the result of a comparison between
// two SchemeVersions.
type Comparison uint8

const (
	LessThan Comparison = iota
	EqualTo
	GreaterThan
)

const (
	tagAddedIn      = "addedIn"
	tagDeprecatedIn = "deprecatedIn"
	tagRemovedIn    = "removedIn"
)

// SchemeVersion is a representation of a configuration scheme version
// that can be compared to other SchemeVersions.
type SchemeVersion struct {
	Major int
	Minor int
}

// NewSchemeVersion creates a new instance of a SchemeVersion.
func NewSchemeVersion(versionString string) (*SchemeVersion, error) {
	var min, maj int
	var err error

	if versionString == "" {
		return nil, fmt.Errorf("no version info found")
	}

	s := strings.Split(versionString, ".")
	if len(s) == 1 {
		maj, err = strconv.Atoi(s[0])
		if err != nil {
			return nil, err
		}
		min = 0
	} else {
		maj, err = strconv.Atoi(s[0])
		if err != nil {
			return nil, err
		}
		min, err = strconv.Atoi(s[1])
		if err != nil {
			return nil, err
		}
	}

	return &SchemeVersion{
		Major: maj,
		Minor: min,
	}, nil
}

// String returns a string representation of the scheme version.
func (schemeVersion *SchemeVersion) String() string {
	return fmt.Sprintf("%d.%d", schemeVersion.Major, schemeVersion.Minor)
}

// Compare compares a SchemeVersion with another SchemeVersion.
//
// If this SchemeVersion has a greater version number than the provided
// one, GreaterThan is returned. If they are equal versions, EqualTo is
// returned. If it is a lower version, LessThan is returned.
func (schemeVersion *SchemeVersion) Compare(other *SchemeVersion) Comparison {
	if schemeVersion.Major > other.Major {
		return GreaterThan
	}

	if schemeVersion.Major < other.Major {
		return LessThan
	}

	if schemeVersion.Minor > other.Minor {
		return GreaterThan
	}

	if schemeVersion.Minor < other.Minor {
		return LessThan
	}
	return EqualTo
}

// ConfigVersion is a struct that is used to extract the configuration
// scheme version from any config file.
type ConfigVersion struct {
	// Version is the config version scheme specified in the config file.
	Version string

	// file is the path of the file that the version was read from.
	file string

	// scheme is the SchemeVersion that represents the ConfigVersion's Version.
	scheme *SchemeVersion
}

// parseScheme parses the Version field into a SchemeVersion.
func (configVersion *ConfigVersion) parseScheme() (err error) {
	var min, maj int

	if configVersion.Version == "" {
		return fmt.Errorf("no version info found")
	}

	s := strings.Split(configVersion.Version, ".")
	if len(s) == 1 {
		maj, err = strconv.Atoi(s[0])
		if err != nil {
			return
		}
		min = 0
	} else {
		maj, err = strconv.Atoi(s[0])
		if err != nil {
			return
		}
		min, err = strconv.Atoi(s[1])
		if err != nil {
			return
		}
	}

	configVersion.scheme = &SchemeVersion{
		Major: maj,
		Minor: min,
	}
	return nil
}

// Validate validates that the ConfigVersion has no configuration errors.
func (configVersion *ConfigVersion) Validate() error {
	v, err := NewSchemeVersion(configVersion.Version)
	if err != nil {
		return err
	}
	configVersion.scheme = v
	return nil
}

// GetSchemeVersion gets the SchemeVersion associated with the version specified
// in the configuration.
func (configVersion *ConfigVersion) GetSchemeVersion() (*SchemeVersion, error) {
	if configVersion.scheme == nil {
		err := configVersion.parseScheme()
		if err != nil {
			return nil, err
		}
	}
	return configVersion.scheme, nil
}
