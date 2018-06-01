package sdk

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vapor-ware/synse-server-grpc/go"

	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
	"golang.org/x/time/rate"
)

// DataManager is the global data manager for the plugin.
var DataManager = &dataManager{
	// Do not make the read/write channel. Those channels will be set up
	// when the DataManger is initialized via `dataManager.init()`
	readings: make(map[string][]*Reading),
	dataLock: &sync.RWMutex{},
	rwLock:   &sync.Mutex{},
}

// dataManager handles the reading from and writing to configured devices.
// It executes the read and write goroutines and uses the channels between
// those goroutines and its process to update the read and write state.
//
// If the plugin is configured to run in "serial" mode, the dataManager
// also manages the locking around access to state across processes.
type dataManager struct {

	// readChannel is the channel that is used to get data from the
	// device being read and update the `readings` field accordingly.
	// Readings are sent to the channel by the `pollRead` function and
	// are received by the `goUpdateData` goroutine.
	readChannel chan *ReadContext

	// writeChannel is the channel that is used to write data to devices.
	// Write data is sent to the channel by the `Write` function and is
	// received by the `pollWrite` function.
	writeChannel chan *WriteContext

	// readings is a map of readings, where the key is the GUID of a
	// device, and the values are the readings associated with that device.
	readings map[string][]*Reading

	// Lock around access/update of the `readings` map data.
	dataLock *sync.RWMutex

	// Lock around async reads and writes.
	rwLock *sync.Mutex

	// limiter is a rate limiter for making requests. This is configured
	// via the plugin config.
	limiter *rate.Limiter
}

// init initializes the data manager by creating the structures needed, based
// on the global plugin configuration, and by string the read, write, and updater
// goroutines, allowing it to provide data from and access to configured devices.
func (manager *dataManager) init() {
	logger.Info("Initializing dataManager goroutines..")

	// Initialize the read and write channels
	manager.readChannel = make(chan *ReadContext, PluginConfig.Settings.Read.Buffer)
	manager.writeChannel = make(chan *WriteContext, PluginConfig.Settings.Write.Buffer)

	// Initialize the limiter, if configured
	if PluginConfig.Limiter != nil && PluginConfig.Limiter != (&config.LimiterSettings{}) {
		manager.limiter = rate.NewLimiter(
			rate.Limit(PluginConfig.Limiter.Rate),
			int(PluginConfig.Limiter.Burst),
		)
	}

	// Start the reader/writer
	manager.goRead()
	manager.goWrite()

	// Update the manager readings state
	manager.goUpdateData()

	logger.Info("dataManager initialization complete.")
}

// writesEnabled checks to see whether writing is enabled for the plugin based on
// the configuration.
func (manager *dataManager) writesEnabled() bool {
	return PluginConfig.Settings.Write.Enabled
}

// goRead starts the goroutine for reading from configured devices.
func (manager *dataManager) goRead() {
	// If reads are not enabled, there is nothing to do here.
	if !PluginConfig.Settings.Read.Enabled {
		logger.Info("plugin reads disabled in config - will not start the read goroutine")
		return
	}

	logger.Info("plugin reads enabled - starting the read goroutine")
	go func() {
		interval, _ := PluginConfig.Settings.Read.GetInterval()
		for {
			// Perform the reads. This is done in a separate function
			// to allow for cleaner lock/unlock semantics.
			switch mode := PluginConfig.Settings.Mode; mode {
			case "serial":
				// Get device readings in serial
				manager.serialRead()
			case "parallel":
				// Get device readings in parallel
				manager.parallelRead()
			default:
				logger.Errorf("exiting read loop: unsupported plugin run mode: %s", mode)
				return
			}

			time.Sleep(interval)
		}
	}()
}

// readOne implements the logic for reading from an individual device that is
// configured with the Plugin.
func (manager *dataManager) readOne(device *Device) {
	// Rate limiting, if configured
	if manager.limiter != nil {
		err := manager.limiter.Wait(context.Background())
		if err != nil {
			logger.Errorf("error from limiter when reading %v: %v", device.GUID(), err)
		}
	}

	// If the device does not get its readings from a bulk read operation,
	// then it is read individually. If a device is read in bulk, it will
	// not be read here; it will be read via the readBulk function.
	if !device.bulkRead {
		resp, err := device.Read()
		if err != nil {
			// Check to see if the error is that of unsupported error. If it is, we
			// do not want to log out here (low-interval read polling would cause this
			// to pollute the logs for something that we should already know).
			_, unsupported := err.(*errors.UnsupportedCommandError)
			if !unsupported {
				logger.Errorf("failed to read from device %v: %v", device.GUID(), err)
			}
		} else {
			manager.readChannel <- resp
		}
	}
}

// readBulk will execute bulk reads on all device handlers that support
// bulk reading. If a handler does not support bulk reading, it's devices
// will be read individually via readOne instead.
func (manager *dataManager) readBulk(handler *DeviceHandler) {
	// Rate limiting, if configured
	if manager.limiter != nil {
		err := manager.limiter.Wait(context.Background())
		if err != nil {
			logger.Errorf("error from limiter when bulk reading with handler for %v: %v", handler.Name, err)
		}
	}

	// If the handler supports bulk read, execute bulk read. Otherwise,
	// do nothing. Individual reads are done via the readOne function.
	if handler.supportsBulkRead() {
		devices := handler.getDevicesForHandler()
		resp, err := handler.BulkRead(devices)
		if err != nil {
			logger.Errorf("failed to bulk read from device handler for: %v: %v", handler.Name, err)
		} else {
			for _, readCtx := range resp {
				manager.readChannel <- readCtx
			}
		}
	}
}

// serialRead reads all devices configured with the Plugin in serial.
func (manager *dataManager) serialRead() {
	// If the plugin is a serial plugin, we want to lock around reads
	// and writes so the two operations do not stomp on one another.
	manager.rwLock.Lock()
	defer manager.rwLock.Unlock()

	for _, dev := range deviceMap {
		manager.readOne(dev)
	}

	for _, handler := range deviceHandlers {
		manager.readBulk(handler)
	}
}

// parallelRead reads all devices configured with the Plugin in parallel.
func (manager *dataManager) parallelRead() {
	var waitGroup sync.WaitGroup

	for _, dev := range deviceMap {
		// Increment the WaitGroup counter.
		waitGroup.Add(1)

		// Launch a goroutine to read from the device
		go func(wg *sync.WaitGroup, device *Device) {
			manager.readOne(device)
			wg.Done()
		}(&waitGroup, dev)
	}

	for _, handler := range deviceHandlers {
		// Increment the WaitGroup counter.
		waitGroup.Add(1)

		// Launch a goroutine to bulk read from the handler
		go func(wg *sync.WaitGroup, handler *DeviceHandler) {
			manager.readBulk(handler)
			wg.Done()
		}(&waitGroup, handler)
	}

	// Wait for all device reads to complete.
	waitGroup.Wait()
}

// goWrite starts the goroutine for writing to configured devices.
func (manager *dataManager) goWrite() {
	// If writes are not enabled, there is nothing to do here.
	if !manager.writesEnabled() {
		logger.Info("plugin writes disabled in config - will not start the write goroutine")
		return
	}

	logger.Info("plugin writes enabled - starting the write goroutine")
	go func() {
		interval, _ := PluginConfig.Settings.Write.GetInterval()
		for {
			// Perform the writes. This is done in a separate function
			// to allow for cleaner lock/unlock semantics.
			switch mode := PluginConfig.Settings.Mode; mode {
			case "serial":
				// Write to devices in serial
				manager.serialWrite()
			case "parallel":
				// Write to devices in parallel
				manager.parallelWrite()
			default:
				logger.Errorf("exiting write loop: unsupported plugin run mode: %s", mode)
				return
			}

			time.Sleep(interval)
		}
	}()
}

// serialWrite writes to devices configured with the Plugin in serial.
func (manager *dataManager) serialWrite() {
	// If the plugin is a serial plugin, we want to lock around reads
	// and writes so the two operations do not stomp on one another.
	manager.rwLock.Lock()
	defer manager.rwLock.Unlock()

	// Check for any pending writes and, if any exist, attempt to fulfill
	// the writes and update their transaction state accordingly.
	for i := 0; i < PluginConfig.Settings.Write.Max; i++ {
		select {
		case w := <-manager.writeChannel:
			manager.write(w)

		default:
			// if there is nothing to write, do nothing
		}
	}
}

// parallelWrite writes to devices configured with the Plugin in parallel.
func (manager *dataManager) parallelWrite() {
	var waitGroup sync.WaitGroup

	// Check for any pending writes and, if any exist, attempt to fulfill
	// the writes and update their transaction state accordingly.
	for i := 0; i < PluginConfig.Settings.Write.Max; i++ {
		select {
		case w := <-manager.writeChannel:
			// Increment the WaitGroup counter.
			waitGroup.Add(1)

			// Launch a goroutine to write to the device
			go func(wg *sync.WaitGroup, writeContext *WriteContext) {
				manager.write(writeContext)
				wg.Done()
			}(&waitGroup, w)

		default:
			// if there is nothing to write, do nothing
		}
	}

	// Wait for all device reads to complete.
	waitGroup.Wait()
}

// write implements the logic for writing to a devices that is configured
// with the Plugin.
func (manager *dataManager) write(w *WriteContext) {
	// Rate limiting, if configured
	if manager.limiter != nil {
		err := manager.limiter.Wait(context.Background())
		if err != nil {
			logger.Errorf("error from limiter when writing %v: %v", w, err)
		}
	}

	// Write to the device
	logger.Debugf("writing for %v (transaction %v)", w.device, w.transaction.id)
	w.transaction.setStatusWriting()

	device := deviceMap[w.ID()]
	if device == nil {
		w.transaction.setStateError()
		msg := "no device found with ID " + w.ID()
		w.transaction.message = msg
		logger.Error(msg)
	}

	data := decodeWriteData(w.data)
	err := device.Write(data)
	if err != nil {
		w.transaction.setStateError()
		w.transaction.message = err.Error()
		logger.Errorf("failed to write to device %v: %v", w.device, err)
	}
	w.transaction.setStatusDone()
}

// goUpdateData updates the DeviceManager's readings state with the latest
// values that were read for each device.
func (manager *dataManager) goUpdateData() {
	logger.Info("starting data updater")
	go func() {
		for {
			reading := <-manager.readChannel
			manager.dataLock.Lock()
			manager.readings[reading.ID()] = reading.Reading
			manager.dataLock.Unlock()
		}
	}()
}

// getReadings safely gets a reading value from the dataManager readings field by
// accessing the readings for the specified device within a lock context. Since the
// readings map is updated in a separate goroutine, we want to lock access around the
// map to prevent simultaneous access collisions.
func (manager *dataManager) getReadings(device string) []*Reading {
	manager.dataLock.RLock()
	defer manager.dataLock.RUnlock()

	return manager.readings[device]
}

// Read fulfills a Read request by providing the latest data read from a device
// and framing it up for the gRPC response.
func (manager *dataManager) Read(req *synse.DeviceFilter) ([]*synse.Reading, error) {
	// Validate that the incoming request has the requisite fields populated.
	err := validateReadRequest(req)
	if err != nil {
		logger.Errorf("Incoming read request failed validation %v: %v", req, err)
		return nil, err
	}

	// Create the id for the device.
	deviceID := makeIDString(req.Rack, req.Board, req.Device)
	err = validateForRead(deviceID)
	if err != nil {
		logger.Errorf("Unable to read device %s: %v", deviceID, err)
		return nil, err
	}

	// Get the readings for the device.
	readings := manager.getReadings(deviceID)
	if readings == nil {
		logger.Errorf("No readings found for device: %s", deviceID)
		return nil, errors.NotFoundErr("no readings found for device: %s", deviceID)
	}

	// Create the response containing the device readings.
	var resp []*synse.Reading
	for _, r := range readings {
		resp = append(resp, r.encode())
	}
	return resp, nil
}

// Write fulfills a Write request by queuing up the write context and framing
// up the corresponding gRPC response.
func (manager *dataManager) Write(req *synse.WriteInfo) (map[string]*synse.WriteData, error) {
	// Validate that the incoming request has the requisite fields populated.
	err := validateWriteRequest(req)
	if err != nil {
		logger.Errorf("Incoming write request failed validation %v: %v", req, err)
		return nil, err
	}

	filter := req.DeviceFilter

	// Create the id for the device.
	deviceID := makeIDString(filter.Rack, filter.Board, filter.Device)
	err = validateForWrite(deviceID)
	if err != nil {
		logger.Errorf("Unable to write to device %s: %v", deviceID, err)
		return nil, err
	}

	// Ensure writes are enabled.
	if enabled := manager.writesEnabled(); !enabled {
		return nil, fmt.Errorf("writing is not enabled")
	}

	// Perform the write and build the response.
	var resp = make(map[string]*synse.WriteData)
	for _, data := range req.Data {
		t := newTransaction()
		t.setStatusPending()

		// Map the transaction ID to the write context for the response
		resp[t.id] = data

		// Pass the write context to the write channel to be queued for writing.
		manager.writeChannel <- &WriteContext{
			transaction: t,
			device:      filter.Device,
			board:       filter.Board,
			rack:        filter.Rack,
			data:        data,
		}
	}
	logger.Debugf("write response data: %#v", resp)
	return resp, nil
}
