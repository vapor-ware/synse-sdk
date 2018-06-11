package outputs

import "github.com/vapor-ware/synse-sdk/sdk/oldconfig"

var (
	// AirflowOutput is the output for airflow devices.
	AirflowOutput = oldconfig.OutputType{
		Name:      "airflow",
		Precision: 2,
		Unit: oldconfig.Unit{
			Name:   "cubic feet per meter",
			Symbol: "CFM",
		},
	}

	// TemperatureOutput is the output for temperature devices.
	TemperatureOutput = oldconfig.OutputType{
		Name:      "temperature",
		Precision: 2,
		Unit: oldconfig.Unit{
			Name:   "celsius",
			Symbol: "C",
		},
	}

	// VoltageOutput is the output for voltage devices.
	VoltageOutput = oldconfig.OutputType{
		Name:      "voltage",
		Precision: 5,
		Unit: oldconfig.Unit{
			Name:   "volts",
			Symbol: "V",
		},
	}
)
