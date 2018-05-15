package sdk

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vapor-ware/synse-server-grpc/go"

	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
)

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

	// Map of readings as strings. Key is the device UID.
	readings map[string][]*Reading

	// Lock around access/update of the `readings` map data.
	dataLock *sync.RWMutex

	// Lock around async reads and writes.
	rwLock *sync.Mutex

	// The plugin's handlers. See sdk/handlers.go.
	handlers *Handlers

	// The plugin's device handlers. These are what handle the read
	// and write functionality for all devices.
	deviceHandlers []*DeviceHandler

	// The plugin's configuration. See sdk.config.plugin.go.
	config *config.PluginConfig
}

// newDataManager creates a new instance of the dataManager for a Plugin. For a
// new dataManager to be created successfully, the given Plugin should be non-nil,
// have non-nil Handlers defined, and have a non-nil config defined.
func newDataManager(plugin *Plugin) (*dataManager, error) {
	// Nil check the parameter and all pointers we dereference here for now.
	// FUTURE: Check in the plugin config constructor.
	if plugin == nil {
		return nil, fmt.Errorf("plugin parameter must not be nil")
	}

	if plugin.handlers == nil {
		return nil, fmt.Errorf("plugin.handlers in parameter must not be nil")
	}

	if plugin.Config == nil {
		return nil, fmt.Errorf("plugin.Config in parameter must not be nil")
	}

	return &dataManager{
		readChannel:    make(chan *ReadContext, plugin.Config.Settings.Read.Buffer),
		writeChannel:   make(chan *WriteContext, plugin.Config.Settings.Write.Buffer),
		readings:       make(map[string][]*Reading),
		dataLock:       &sync.RWMutex{},
		rwLock:         &sync.Mutex{},
		handlers:       plugin.handlers,
		deviceHandlers: plugin.deviceHandlers,
		config:         plugin.Config,
	}, nil
}

// init initializes the goroutines for the dataManager so it can start reading
// from and writing to the devices managed by the Plugin.
func (manager *dataManager) init() {
	logger.Info("Initializing dataManager goroutines..")

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
	return manager.config.Settings.Write.Enabled
}

// goRead starts the goroutine for reading from configured devices.
func (manager *dataManager) goRead() {
	// If reads are not enabled, there is nothing to do here.
	if !manager.config.Settings.Read.Enabled {
		logger.Info("plugin reads disabled in config - will not start the read goroutine")
		return
	}

	logger.Info("plugin reads enabled - starting the read goroutine")
	go func() {
		interval, _ := manager.config.Settings.Read.GetInterval()
		for {
			// Perform the reads. This is done in a separate function
			// to allow for cleaner lock/unlock semantics.
			switch mode := manager.config.Settings.Mode; mode {
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

// read implements the logic for reading from a device that is configured
// with the Plugin.
func (manager *dataManager) readOne(device *Device) {
	// Rate limiting, if configured
	if manager.config.Limiter != nil {
		err := manager.config.Limiter.Wait(context.Background())
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
			_, unsupported := err.(*UnsupportedCommandError)
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
	if manager.config.Limiter != nil {
		err := manager.config.Limiter.Wait(context.Background())
		if err != nil {
			logger.Errorf("error from limiter when bulk reading with handler for %v: %v", handler.Model, err)
		}
	}

	// If the handler supports bulk read, execute bulk read. Otherwise,
	// do nothing. Individual reads are done via the readOne function.
	if handler.doesBulkRead() {
		devices := handler.getDevicesForHandler()
		resp, err := handler.BulkRead(devices)
		if err != nil {
			logger.Errorf("failed to bulk read from device handler for: %v: %v", handler.Model, err)
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

	for _, handler := range manager.deviceHandlers {
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

	for _, handler := range manager.deviceHandlers {
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
	if !manager.config.Settings.Write.Enabled {
		logger.Info("plugin writes disabled in config - will not start the write goroutine")
		return
	}

	logger.Info("plugin writes enabled - starting the write goroutine")
	go func() {
		interval, _ := manager.config.Settings.Write.GetInterval()
		for {
			// Perform the writes. This is done in a separate function
			// to allow for cleaner lock/unlock semantics.
			switch mode := manager.config.Settings.Mode; mode {
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

func (manager *dataManager) serialWrite() {
	// If the plugin is a serial plugin, we want to lock around reads
	// and writes so the two operations do not stomp on one another.
	manager.rwLock.Lock()
	defer manager.rwLock.Unlock()

	// Check for any pending writes and, if any exist, attempt to fulfill
	// the writes and update their transaction state accordingly.
	for i := 0; i < manager.config.Settings.Write.Max; i++ {
		select {
		case w := <-manager.writeChannel:
			manager.write(w)

		default:
			// if there is nothing to write, do nothing
		}
	}
}

func (manager *dataManager) parallelWrite() {
	var waitGroup sync.WaitGroup

	// Check for any pending writes and, if any exist, attempt to fulfill
	// the writes and update their transaction state accordingly.
	for i := 0; i < manager.config.Settings.Write.Max; i++ {
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
	if manager.config.Limiter != nil {
		err := manager.config.Limiter.Wait(context.Background())
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
func (manager *dataManager) Read(req *synse.ReadRequest) ([]*synse.ReadResponse, error) {
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
		return nil, notFoundErr("no readings found for device: %s", deviceID)
	}

	// Create the response containing the device readings.
	var resp []*synse.ReadResponse
	for _, r := range readings {
		reading := &synse.ReadResponse{
			Timestamp: r.Timestamp,
			Type:      r.Type,
			Value:     r.Value,
		}
		resp = append(resp, reading)
	}
	return resp, nil
}

// Write fulfills a Write request by queuing up the write context and framing
// up the corresponding gRPC response.
func (manager *dataManager) Write(req *synse.WriteRequest) (map[string]*synse.WriteData, error) {
	// Validate that the incoming request has the requisite fields populated.
	err := validateWriteRequest(req)
	if err != nil {
		logger.Errorf("Incoming write request failed validation %v: %v", req, err)
		return nil, err
	}

	// Create the id for the device.
	deviceID := makeIDString(req.Rack, req.Board, req.Device)
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
		t, err := newTransaction()
		if err != nil {
			return nil, err
		}
		t.setStatusPending()

		resp[t.id] = data
		manager.writeChannel <- &WriteContext{
			transaction: t,
			device:      req.Device,
			board:       req.Board,
			rack:        req.Rack,
			data:        data,
		}
	}
	logger.Debugf("write response data: %#v", resp)
	return resp, nil
}
