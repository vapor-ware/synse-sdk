package cfg

// ConfigComponent is an interface that all structs that define configuration
// components should implement.
//
// This interface implements a Validate function which is used by the
// SchemeValidator in order to validate each struct that makes up a configuration.
type ConfigComponent interface {
	Validate() error
}
