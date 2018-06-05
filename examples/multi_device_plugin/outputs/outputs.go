package outputs

import "github.com/vapor-ware/synse-sdk/sdk/config"

var (
	// AirflowOutput is the output for airflow devices.
	AirflowOutput = config.OutputType{
		Name:      "airflow",
		Precision: 2,
		Unit: config.Unit{
			Name:   "cubic feet per meter",
			Symbol: "CFM",
		},
	}

	// TemperatureOutput is the output for temperature devices.
	TemperatureOutput = config.OutputType{
		Name:      "temperature",
		Precision: 2,
		Unit: config.Unit{
			Name:   "celsius",
			Symbol: "C",
		},
	}

	// VoltageOutput is the output for voltage devices.
	VoltageOutput = config.OutputType{
		Name:      "voltage",
		Precision: 5,
		Unit: config.Unit{
			Name:   "volts",
			Symbol: "V",
		},
	}
)
