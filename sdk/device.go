package sdk

import (
	"github.com/vapor-ware/synse-sdk/sdk/types"
)

// NDevice is a New Device. It will replace Device once v1.0 is complete.
type NDevice struct {
	Location Location

	Kind string

	Metadata map[string]string

	Plugin string

	Info string

	Data map[string]interface{}

	Outputs []*Output

	id       string
	bulkRead bool
}

// fixme: unclear if we need this or if we just want to use the
// struct types defined in cfg...
type Location struct {
	Rack  string
	Board string
}

type Output struct {
	types.ReadingType

	Info string
	Data map[string]interface{}
}
