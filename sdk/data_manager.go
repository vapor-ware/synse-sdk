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

// DataManager handles the reading from and writing to configured devices.
type DataManager struct {
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

	// The plugin's device handlers. See sdk/handlers.go.
	handlers *Handlers

	// The plugin's configuration. See sdk.config.plugin.go.
	config *config.PluginConfig
}

// NewDataManager creates a new instance of the DataManager using the existing
// configurations and handlers registered with the plugin.
func NewDataManager(plugin *Plugin) (*DataManager, error) {
	// Nil check the parameter and all pointers we dereference here for now.
	// FUTURE: Check in the plugin config constructor.
	// FIXME (etd) - I don't think we'd necessarily want invalidArgumentErr
	// here -- that is used for gRPC errors. This is just a plugin error that
	// should terminate the plugin (misconfigured/misconstructed).
	if plugin == nil {
		return nil, invalidArgumentErr("plugin parameter must not be nil")
	}

	if plugin.handlers == nil {
		return nil, invalidArgumentErr("plugin.handlers in parameter must not be nil")
	}

	if plugin.Config == nil {
		return nil, invalidArgumentErr("plugin.Config in parameter must not be nil")
	}

	return &DataManager{
		// TODO: https://github.com/vapor-ware/synse-sdk/issues/118
		readChannel:  make(chan *ReadContext, plugin.Config.Settings.Read.Buffer),
		writeChannel: make(chan *WriteContext, plugin.Config.Settings.Write.Buffer),
		readings:     make(map[string][]*Reading),
		dataLock:     &sync.RWMutex{},
		rwLock:       &sync.Mutex{},
		handlers:     plugin.handlers,
		config:       plugin.Config,
	}, nil
}

// init initializes the goroutines for the DataManager so it can start reading
// from and writing to the devices managed by the Plugin.
func (manager *DataManager) init() {
	logger.Info("Initializing DataManager goroutines..")

	// Start the reader/writer
	manager.goRead()
	manager.goWrite()

	// Update the manager readings state
	manager.goUpdateData()

	logger.Info("DataManager initialization complete.")
}

// writesEnabled checks to see whether writing is enabled for the plugin based on
// the configuration.
func (manager *DataManager) writesEnabled() bool {
	return manager.config.Settings.Write.Enabled
}

// goRead starts the goroutine for reading from configured devices.
func (manager *DataManager) goRead() {
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
func (manager *DataManager) read(device *Device) {
	// Rate limiting, if configured
	if manager.config.Limiter != nil {
		err := manager.config.Limiter.Wait(context.Background())
		if err != nil {
			logger.Errorf("error from limiter when reading %v: %v", device.GUID(), err)
		}
	}

	// Read from the device
	resp, err := device.Read()
	if err != nil {
		logger.Errorf("failed to read from device %v: %v", device.GUID(), err)
	} else {
		manager.readChannel <- resp
	}
}

// serialRead reads all devices configured with the Plugin in serial.
func (manager *DataManager) serialRead() {
	// If the plugin is a serial plugin, we want to lock around reads
	// and writes so the two operations do not stomp on one another.
	manager.rwLock.Lock()
	defer manager.rwLock.Unlock()

	for _, dev := range deviceMap {
		manager.read(dev)
	}
}

// parallelRead reads all devices configured with the Plugin in parallel.
func (manager *DataManager) parallelRead() {
	var waitGroup sync.WaitGroup

	for _, dev := range deviceMap {
		// Increment the WaitGroup counter.
		waitGroup.Add(1)

		// Launch a goroutine to read from the device
		go func(wg *sync.WaitGroup, device *Device) {
			manager.read(device)
			wg.Done()
		}(&waitGroup, dev)
	}

	// Wait for all device reads to complete.
	waitGroup.Wait()
}

// goWrite starts the goroutine for writing to configured devices.
func (manager *DataManager) goWrite() {
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

func (manager *DataManager) serialWrite() {
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

func (manager *DataManager) parallelWrite() {
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
func (manager *DataManager) write(w *WriteContext) {
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
func (manager *DataManager) goUpdateData() {
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

// getReadings safely gets a reading value from the DataManager readings field by
// accessing the readings for the specified device within a lock context. Since the
// readings map is updated in a separate goroutine, we want to lock access around the
// map to prevent simultaneous access collisions.
func (manager *DataManager) getReadings(device string) []*Reading {
	manager.dataLock.RLock()
	defer manager.dataLock.RUnlock()

	return manager.readings[device]
}

// Read fulfills a Read request by providing the latest data read from a device
// and framing it up for the gRPC response.
func (manager *DataManager) Read(req *synse.ReadRequest) ([]*synse.ReadResponse, error) {
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
func (manager *DataManager) Write(req *synse.WriteRequest) (map[string]*synse.WriteData, error) {
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
		t, err := NewTransaction()
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
