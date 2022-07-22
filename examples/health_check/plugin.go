package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/v2/sdk"
	"github.com/vapor-ware/synse-sdk/v2/sdk/health"
	"github.com/vapor-ware/synse-sdk/v2/sdk/output"
)

var (
	pluginName       = "health check plugin"
	pluginMaintainer = "vaporio"
	pluginDesc       = "A example plugin with health checks"
)

var (
	// temperatureHandler defines the read/write behavior for the "example.temperature" device kind.
	temperatureHandler = sdk.DeviceHandler{
		Name: "example.temperature",

		Read: func(device *sdk.Device) ([]*output.Reading, error) {
			reading, err := output.Temperature.MakeReading(strconv.Itoa(rand.Int()))
			if err != nil {
				return nil, err
			}
			return []*output.Reading{reading}, nil
		},
		Write: func(device *sdk.Device, data *sdk.WriteData) error {
			fmt.Printf("[temperature handler]: WRITE (%v)\n", device.GetID())
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
		log.Error("[health check]: error")
		return fmt.Errorf("error file detected")
	}
	log.Info("[health check]: ok")
	return nil
}

func main() {
	// Set the metadata for the plugin.
	sdk.SetPluginInfo(
		pluginName,
		pluginMaintainer,
		pluginDesc,
		"",
	)

	// Create a new Plugin instance.
	plugin, err := sdk.NewPlugin()
	if err != nil {
		log.Fatal(err)
	}

	// Register our device handlers with the Plugin.
	err = plugin.RegisterDeviceHandlers(
		&temperatureHandler,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register the health check with the plugin
	customCheck := health.NewPeriodicHealthCheck("example health check", 3*time.Second, CustomHealthCheck)
	if err := plugin.RegisterHealthChecks(customCheck); err != nil {
		log.Fatal(err)
	}

	// Run the plugin.
	if err := plugin.Run(); err != nil {
		log.Fatal(err)
	}
}
