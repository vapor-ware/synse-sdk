// Synse SDK
// Copyright (c) 2019-2020 Vapor IO
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package output

// GetBuiltins returns a list of all the built-in outputs supplied by the SDK.
func GetBuiltins() []*Output {
	return []*Output{
		&Color,
		&Direction,
		&ElectricCurrent,
		&ElectricResistance,
		&Frequency,
		&Humidity,
		&Kilojoule,
		&KilowattHour,
		&Pascal,
		&Percentage,
		&PSI,
		&RPM,
		&Seconds,
		&State,
		&Status,
		&Temperature,
		&Velocity,
		&Voltage,
		&Watt,
		&Weber,
	}
}

// Color is an output type for color readings. This output has no unit.
//
// A color reading is generally a string which represents some kind of
// color. This can be the name ("red"), a hex string ("ff0000"), an
// RBG string ("255,0,0"), or anything else.
var Color = Output{
	Name: "color",
	Type: "color",
}

// Direction is an output type for directionality (e.g. the direction
// in which a fan may be spinning). This output has no unit.
var Direction = Output{
	Name: "direction",
	Type: "direction",
}

// ElectricCurrent is an output type for electrical current readings,
// measured in Amperes.
var ElectricCurrent = Output{
	Name:      "electric-current",
	Type:      "current",
	Precision: 3,
	Unit: &Unit{
		Name:   "ampere",
		Symbol: "A",
	},
}

// ElectricResistance is an output type for electrical resistance readings,
// measured in Ohms.
var ElectricResistance = Output{
	Name:      "electric_resistance",
	Type:      "resistance",
	Precision: 2,
	Unit: &Unit{
		Name:   "ohm",
		Symbol: "Ω",
	},
}

// Frequency is an output type for frequency readings, measured in Hertz.
var Frequency = Output{
	Name:      "frequency",
	Type:      "frequency",
	Precision: 2,
	Unit: &Unit{
		Name:   "hertz",
		Symbol: "Hz",
	},
}

// Humidity is an output type for humidity readings, measured as a percentage.
var Humidity = Output{
	Name:      "humidity",
	Type:      "humidity",
	Precision: 2,
	Unit: &Unit{
		Name:   "percent humidity",
		Symbol: "%",
	},
}

// Kilojoule is an output type for energy readings, measured in kilojoules.
var Kilojoule = Output{
	Name:      "kilojoule",
	Type:      "energy",
	Precision: 3,
	Unit: &Unit{
		Name:   "kilojoule",
		Symbol: "kJ",
	},
}

// KilowattHour is an output type for energy readings, measured in kilowatt-hours.
var KilowattHour = Output{
	Name:      "kilowatt-hour",
	Type:      "energy",
	Precision: 3,
	Unit: &Unit{
		Name:   "kilowatt-hour",
		Symbol: "kWh",
	},
}

// Pascal is an output type for pressure readings, measured in Pascals.
var Pascal = Output{
	Name:      "pascal",
	Type:      "pressure",
	Precision: 3,
	Unit: &Unit{
		Name:   "pascal",
		Symbol: "Pa",
	},
}

// Percentage is an output type for general percentages.
var Percentage = Output{
	Name: "percentage",
	Type: "percentage",
	Unit: &Unit{
		Name:   "percent",
		Symbol: "%",
	},
}

// PSI is an output type for pressure readings, measured in PSI.
var PSI = Output{
	Name:      "psi",
	Type:      "pressure",
	Precision: 3,
	Unit: &Unit{
		Name:   "pounds per square inch",
		Symbol: "psi",
	},
}

// RPM is an output type for frequency readings, measured in revolutions per minute.
var RPM = Output{
	Name:      "rpm",
	Type:      "frequency",
	Precision: 2,
	Unit: &Unit{
		Name:   "revolutions per minute",
		Symbol: "RPM",
	},
}

// Seconds is an output type for a duration, measured in seconds.
var Seconds = Output{
	Name:      "seconds",
	Type:      "duration",
	Precision: 3,
	Unit: &Unit{
		Name:   "seconds",
		Symbol: "s",
	},
}

// State is an output type for state readings. This output has no unit.
//
// A state reading is generally a string which represents some kind of
// state (e.g. "on"/"off").
var State = Output{
	Name: "state",
	Type: "state",
}

// Status is an output type for status readings. This output has no unit.
//
// A status reading is generally a string which describes the status of
// a device, e.g. "operational".
var Status = Output{
	Name: "status",
	Type: "status",
}

// Temperature is an output type for temperature readings, measured in degrees
// Celsius.
var Temperature = Output{
	Name:      "temperature",
	Type:      "temperature",
	Precision: 2,
	Unit: &Unit{
		Name:   "celsius",
		Symbol: "C",
	},
}

// Velocity is an output type for velocity readings, measured in meters per second.
var Velocity = Output{
	Name:      "velocity",
	Type:      "velocity",
	Precision: 3,
	Unit: &Unit{
		Name:   "meters per second",
		Symbol: "m/s",
	},
}

// Voltage is an output type for voltage readings, measured in volts.
var Voltage = Output{
	Name:      "voltage",
	Type:      "voltage",
	Precision: 5,
	Unit: &Unit{
		Name:   "volt",
		Symbol: "V",
	},
}

// Watt is an output type for power readings, measured in watts.
var Watt = Output{
	Name:      "watt",
	Type:      "power",
	Precision: 3,
	Unit: &Unit{
		Name:   "watt",
		Symbol: "W",
	},
}

// Weber is an output type for magnetic flux readings, measured in Webers.
//
// Note that:
//  Wb = kg•m²/s²•A = Ω•C = V•s = H•A = T•m² = J/A
//
// That means this output can be used in place of a Volt-second, for example.
var Weber = Output{
	Name:      "weber",
	Type:      "flux",
	Precision: 3,
	Unit: &Unit{
		Name:   "weber",
		Symbol: "Wb",
	},
}
