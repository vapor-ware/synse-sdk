package outputs

import (
	"github.com/vapor-ware/synse-sdk/sdk"
)

var (
	// AirflowOutput is the output for airflow devices.
	AirflowOutput = sdk.OutputType{
		Name:      "airflow",
		Precision: 2,
		Unit: sdk.Unit{
			Name:   "cubic feet per meter",
			Symbol: "CFM",
		},
	}

	// TemperatureOutput is the output for temperature devices.
	TemperatureOutput = sdk.OutputType{
		Name:      "temperature",
		Precision: 2,
		Unit: sdk.Unit{
			Name:   "celsius",
			Symbol: "C",
		},
	}

	// VoltageOutput is the output for voltage devices.
	VoltageOutput = sdk.OutputType{
		Name:      "voltage",
		Precision: 5,
		Unit: sdk.Unit{
			Name:   "volts",
			Symbol: "V",
		},
	}
)
