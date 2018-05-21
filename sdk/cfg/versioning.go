package cfg

// configVersion is a struct that is used to extract the configuration
// scheme version from any config file.
type configVersion struct {
	// Version is the config version scheme specified in the config file.
	Version string

	// file is the path of the file that the version was read from.
	file string
}
