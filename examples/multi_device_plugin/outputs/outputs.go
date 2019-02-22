package outputs

import (
	"github.com/vapor-ware/synse-sdk/sdk/output"
)

var (
	// AirflowOutput is the custom output for airflow devices.
	AirflowOutput = output.Output{
		Name:      "airflow",
		Type:      "airflow",
		Precision: 2,
		Units: map[output.SystemOfMeasure]*output.Unit{
			output.NONE: {
				Name:   "cubic feet per meter",
				Symbol: "CFM",
				System: string(output.NONE),
			},
		},
	}
)
