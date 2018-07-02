package errors

import "fmt"

// ConfigsNotFound is an error used when the search for a config file
// results in that file not being found.
type ConfigsNotFound struct {
	// searchPaths is the list of locations where the file was searched for.
	searchPaths []string
}

// NewConfigsNotFoundError returns a new instance of a ConfigsNotFound error.
func NewConfigsNotFoundError(searchPaths []string) *ConfigsNotFound {
	return &ConfigsNotFound{
		searchPaths: searchPaths,
	}
}

// Error returns the error string and fulfils the error interface.
func (e *ConfigsNotFound) Error() string {
	return fmt.Sprintf("no configuration file(s) found in: %s", e.searchPaths)
}
