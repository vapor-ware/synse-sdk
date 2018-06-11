package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"

	"os"
	"time"

	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/health"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
)

var (
	pluginName       = "Health Check Plugin"
	pluginMaintainer = "Vapor IO"
	pluginDesc       = "A example plugin with health checks"
)

var (
	// The output for temperature devices.
	temperatureOutput = sdk.OutputType{
		Name:      "simple.temperature",
		Precision: 2,
		Unit: sdk.Unit{
			Name:   "celsius",
			Symbol: "C",
		},
	}

	// temperatureHandler defines the read/write behavior for the "example.temperature" device kind.
	temperatureHandler = sdk.DeviceHandler{
		Name: "example.temperature",

		Read: func(device *sdk.Device) ([]*sdk.Reading, error) {
			return []*sdk.Reading{
				device.GetOutput("simple.temperature").MakeReading(
					strconv.Itoa(rand.Int()), // nolint: gas
				),
			}, nil
		},
		Write: func(device *sdk.Device, data *sdk.WriteData) error {
			fmt.Printf("[temperature handler]: WRITE (%v)\n", device.ID())
			fmt.Printf("Data   -> %v\n", data.Data)
			fmt.Printf("Action -> %v\n", data.Action)
			return nil
		},
	}
)

// CustomHealthCheck is a health check that this plugin defines. A health check should
// have the function signature `func() error`. If nil is returned, the health check
// is considered OK. If an error is returned, the check is considered failing.
//
// We will run this check periodically to see whether or not an `error` file exists in
// the current directory. This way, when the plugin is running, you can create the
// error file and watch the health check log change.
func CustomHealthCheck() error {
	if _, err := os.Stat("error"); err == nil {
		logger.Error("[health check]: error")
		return fmt.Errorf("error file detected")
	}
	logger.Info("[health check]: ok")
	return nil
}

func main() {
	// Set the metadata for the plugin.
	sdk.SetPluginMeta(
		pluginName,
		pluginMaintainer,
		pluginDesc,
		"",
	)

	// Create a new Plugin instance.
	plugin := sdk.NewPlugin()

	// Register our output types with the Plugin.
	err := plugin.RegisterOutputTypes(
		&temperatureOutput,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register our device handlers with the Plugin.
	plugin.RegisterDeviceHandlers(
		&temperatureHandler,
	)

	// Register the health check with the health catalog
	health.RegisterPeriodicCheck("example health check", 3*time.Second, CustomHealthCheck)

	// Run the plugin.
	if err := plugin.Run(); err != nil {
		log.Fatal(err)
	}
}
