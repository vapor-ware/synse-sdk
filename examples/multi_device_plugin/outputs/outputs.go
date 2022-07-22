package outputs

import (
	"github.com/vapor-ware/synse-sdk/v2/sdk/output"
)

var (
	// AirflowOutput is the custom output for airflow devices.
	AirflowOutput = output.Output{
		Name:      "airflow",
		Type:      "airflow",
		Precision: 2,
		Unit: &output.Unit{
			Name:   "cubic feet per meter",
			Symbol: "CFM",
		},
	}
)
