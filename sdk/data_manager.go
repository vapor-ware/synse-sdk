package sdk

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-server-grpc/go"
	"golang.org/x/time/rate"
)

// DataManager is the global data manager for the plugin.
var DataManager = newDataManager()

// ListenerCtx is the context needed for a listener function to be called
// and retried at a later time if it errors out after the listener goroutine
// is initially dispatched.
type ListenerCtx struct {
	// handler is the DeviceHandler that defines the handler function.
	handler *DeviceHandler

	// device is the Device that is being listened to via the listener.
	device *Device

	// restarts is the number of times the listener has been restarted.
	restarts int
}

// NewListenerCtx creates a new ListenerCtx for the given handler and device.
func NewListenerCtx(handler *DeviceHandler, device *Device) *ListenerCtx {
	return &ListenerCtx{
		handler:  handler,
		device:   device,
		restarts: 0,
	}
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

	// listenChannel is the channel that is used to get data from a device
	// that is being listened to. This is used by the SDK to collect push
	// based data. While the data in the listenChannel and the data in the
	// readChannel are similar, they are kept separate so the behaviors of
	// pull-based/push-based reading can be tuned independently.
	listenChannel chan *ReadContext

	// listenerRetry is a channel that all listeners will pass a ListenerRetryCtx
	// to if they fail. This channel is read by a separate goroutine which will
	// attempt to re-run the listener.
	listenerRetry chan *ListenerCtx

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

func newDataManager() *dataManager {
	return &dataManager{
		// Do not make the read/write channel. Those channels will be set up
		// when the DataManger is initialized via `dataManager.init()`
		readings: make(map[string][]*Reading),
		dataLock: &sync.RWMutex{},
		rwLock:   &sync.Mutex{},
	}
}

// run sets up the dataManager and starts the read, write, and updater goroutines
// allowing it to provide data from, and access to, configured devices.
func (manager *dataManager) run() error {

	err := manager.setup()
	if err != nil {
		return err
	}

	// Start the listeners/reader/writer
	manager.goListen()
	manager.goRead()
	manager.goWrite()

	// Update the manager readings state
	manager.goUpdateData()

	// Watch for failed listeners to retry them
	go manager.watchForListenerRetry()

	log.Info("[data manager] running")
	return nil
}

// setup initializes the remaining data manager structures based on the global
// plugin configuration.
func (manager *dataManager) setup() error {
	log.Debug("[data manager] setting up data manager state")

	if Config.Plugin == nil {
		return fmt.Errorf("plugin config not set, cannot setup data manager")
	}

	// Initialize the listen, read, and write channels
	manager.listenerRetry = make(chan *ListenerCtx, 50)
	manager.listenChannel = make(chan *ReadContext, Config.Plugin.Settings.Listen.Buffer)
	manager.readChannel = make(chan *ReadContext, Config.Plugin.Settings.Read.Buffer)
	manager.writeChannel = make(chan *WriteContext, Config.Plugin.Settings.Write.Buffer)

	// Initialize the limiter, if configured
	if Config.Plugin.Limiter != nil && Config.Plugin.Limiter != (&LimiterSettings{}) {
		manager.limiter = rate.NewLimiter(
			rate.Limit(Config.Plugin.Limiter.Rate),
			Config.Plugin.Limiter.Burst,
		)
	}
	return nil
}

// writesEnabled checks to see whether writing is enabled for the plugin based on
// the configuration.
func (manager *dataManager) writesEnabled() bool {
	return Config.Plugin.Settings.Write.Enabled
}

// goListen starts the goroutines for any listener functions for the configured
// devices. If there are no listener functions defined, this will do nothing.
func (manager *dataManager) goListen() {
	// Although we consider listening to be a type of "read" behavior (e.g. collecting
	// push-based readings vs. collecting pull-based readings), we use different
	// configuration fields for listening to make it easier to tune independent of
	// pull-based collection needs. If listening is globally disabled, there is
	// nothing to do here.
	if !Config.Plugin.Settings.Listen.Enabled {
		log.Info("[data manager] skipping listener goroutine(s) (listen disabled)")
		return
	}

	// For each handler that has a listener function defined, get the devices for
	// that handler and start the listener for the devices.
	for _, handler := range ctx.deviceHandlers {
		hlog := log.WithField("handler", handler.Name)
		if handler.Listen != nil {
			hlog.Info("[data manager] setting up listeners")

			// Get all of the devices that have registered with the handler
			devices := handler.getDevicesForHandler()
			if len(devices) == 0 {
				hlog.Debugf("[data manager] found no devices for handler")
				continue
			}

			// For each device, run the listener goroutine
			for _, device := range devices {
				ctx := NewListenerCtx(handler, device)
				go manager.runListener(ctx)
			}
		}
	}
}

// runListener runs the listener function for a device. If the listener
// fails, it will attempt to restart the listener.
func (manager *dataManager) runListener(ctx *ListenerCtx) {
	log.WithFields(log.Fields{
		"handler": ctx.handler.Name,
		"device":  ctx.device.ID(),
	}).Info("[data manager] running listener")

	err := ctx.handler.Listen(ctx.device, manager.listenChannel)
	if err != nil {
		log.WithField("device", ctx.device.ID()).Errorf(
			"[data manager] failed to listen for device readings: %v", err,
		)
		// pass the context to retry channel
		manager.listenerRetry <- ctx
	}
}

// watchForListenerRetry waits for the 'runListener' function to pass a
// listener context to it via the 'listenerRetry' channel. If it gets
// a context, that listener had failed and needs to be restarted.
func (manager *dataManager) watchForListenerRetry() {
	for {
		ctx := <-manager.listenerRetry
		// increment the restart counter
		ctx.restarts++

		llog := log.WithFields(log.Fields{
			"manager": ctx.handler.Name,
			"device":  ctx.device.ID(),
		})
		llog.Infof("[data manager] restarting failed listener (restarts %v)", ctx.restarts)
		go manager.runListener(ctx)
	}
}

// goRead starts the goroutine for reading from configured devices.
func (manager *dataManager) goRead() {
	mode := Config.Plugin.Settings.Mode
	readLog := log.WithField("mode", mode)

	// If reads are not enabled, there is nothing to do here.
	if !Config.Plugin.Settings.Read.Enabled {
		readLog.Info("[data manager] skipping read goroutine (reads disabled)")
		return
	}

	readLog.Info("[data manager] starting read goroutine (reads enabled)")
	go func() {
		interval, err := Config.Plugin.Settings.Read.GetInterval()
		if err != nil {
			readLog.WithField("error", err).
				Warn("[data manager] misconfiguration: failed to get read interval")
		}
		for {
			// Perform the reads. This is done in a separate function
			// to allow for cleaner lock/unlock semantics.
			log.Infof("Starting reads in mode %v", mode)
			switch mode {
			case "serial":
				// Get device readings in serial
				serialReadInterval, err := Config.Plugin.Settings.Read.GetSerialReadInterval()
				if err != nil {
					readLog.WithField("error", err).
						Warn("[data manager] misconfiguration: failed to get serial read interval")
				}
				manager.serialRead(serialReadInterval)
			case "parallel":
				// Get device readings in parallel
				manager.parallelRead()
			default:
				readLog.Error("[data manager] exiting read loop: unsupported plugin run mode")
				return
			}

			log.Infof("Completed reads in mode %v", mode)
			log.Infof("Sleeping for interval %v", interval)
			time.Sleep(interval)
			log.Infof("Slept for interval %v", interval)
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
			log.Errorf("[data manager] error from limiter when reading %v: %v", device.GUID(), err)
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
				log.Errorf("[data manager] failed to read from device %v: %v", device.GUID(), err)
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
			log.Errorf("[data manager] error from limiter when bulk reading with handler for %v: %v", handler.Name, err)
		}
	}

	// If the handler supports bulk read, execute bulk read. Otherwise,
	// do nothing. Individual reads are done via the readOne function.
	if handler.supportsBulkRead() {
		devices := handler.getDevicesForHandler()
		if len(devices) == 0 {
			return
		}
		resp, err := handler.BulkRead(devices)
		if err != nil {
			log.Errorf("[data manager] failed to bulk read from device handler for: %v: %v", handler.Name, err)
		} else {
			for _, readCtx := range resp {
				manager.readChannel <- readCtx
			}
		}
	}
}

// serialRead reads all devices configured with the Plugin in serial.
func (manager *dataManager) serialRead(serialReadInterval time.Duration) {
	// If the plugin is a serial plugin, we want to lock around reads
	// and writes so the two operations do not stomp on one another.
	manager.rwLock.Lock()
	defer manager.rwLock.Unlock()

	log.Infof("Starting serial read of %v devices", len(ctx.devices))
	for _, dev := range ctx.devices {
		manager.readOne(dev)
		log.Infof("Sleeping after read %v", serialReadInterval)
		time.Sleep(serialReadInterval)
	}
	log.Infof("Completed serial read of %v devices", len(ctx.devices))

	for _, handler := range ctx.deviceHandlers {
		manager.readBulk(handler)
	}
}

// parallelRead reads all devices configured with the Plugin in parallel.
func (manager *dataManager) parallelRead() {
	var waitGroup sync.WaitGroup

	for _, dev := range ctx.devices {
		// Increment the WaitGroup counter.
		waitGroup.Add(1)

		// Launch a goroutine to read from the device
		go func(wg *sync.WaitGroup, device *Device) {
			manager.readOne(device)
			wg.Done()
		}(&waitGroup, dev)
	}

	for _, handler := range ctx.deviceHandlers {
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
	mode := Config.Plugin.Settings.Mode
	writeLog := log.WithField("mode", mode)

	// If writes are not enabled, there is nothing to do here.
	if !manager.writesEnabled() {
		writeLog.Info("[data manager] skipping write goroutine (writes disabled)")
		return
	}

	writeLog.Info("[data manager] starting write goroutine (writes enabled)")
	go func() {
		interval, err := Config.Plugin.Settings.Write.GetInterval()
		if err != nil {
			writeLog.WithField("error", err).
				Warn("[data manager] misconfiguration: failed to get write interval")
		}
		for {
			// Perform the writes. This is done in a separate function
			// to allow for cleaner lock/unlock semantics.
			switch mode := Config.Plugin.Settings.Mode; mode {
			case "serial":
				// Write to devices in serial
				manager.serialWrite()
			case "parallel":
				// Write to devices in parallel
				manager.parallelWrite()
			default:
				writeLog.Error("[data manager] exiting write loop: unsupported plugin run mode")
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
	for i := 0; i < Config.Plugin.Settings.Write.Max; i++ {
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
	for i := 0; i < Config.Plugin.Settings.Write.Max; i++ {
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
			log.Errorf("[data manager] error from limiter when writing %v: %v", w, err)
		}
	}

	// Write to the device
	log.WithFields(log.Fields{
		"device":      w.device,
		"transaction": w.transaction.id,
	}).Debug("[data manager] fulfilling write transaction")
	w.transaction.setStatusWriting()

	device := ctx.devices[w.ID()]
	if device == nil {
		w.transaction.setStateError()
		msg := "no device found with ID " + w.ID()
		w.transaction.message = msg
		log.Error(msg)
	} else {
		data := decodeWriteData(w.data)
		err := device.Write(data)
		if err != nil {
			w.transaction.setStateError()
			w.transaction.message = err.Error()
			log.Errorf("[data manager] failed to write to device %v: %v", w.device, err)
		}
	}
	w.transaction.setStatusDone()
}

// goUpdateData updates the DeviceManager's readings state with the latest
// values that were read for each device.
func (manager *dataManager) goUpdateData() {
	go func() {
		for {
			var (
				id       string
				readings []*Reading
			)

			// Read from the listen and read channel for incoming readings
			var reading *ReadContext
			select {
			case reading = <-manager.readChannel:
				id = reading.ID()
				readings = reading.Reading
			case reading = <-manager.listenChannel:
				id = reading.ID()
				readings = reading.Reading
			}

			// Update the internal map of current reading state
			manager.dataLock.Lock()
			manager.readings[id] = readings
			manager.dataLock.Unlock()

			// update the readings cache
			addReadingToCache(reading)
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

// getAllReadings safely copies the current reading state in the data manager and
// returns all of the readings.
func (manager *dataManager) getAllReadings() map[string][]*Reading {
	mapCopy := make(map[string][]*Reading)
	manager.dataLock.RLock()
	defer manager.dataLock.RUnlock()

	// Iterate over the map to make a copy - we want a copy or else we would be
	// returning a reference to the underlying data which should only be accessed
	// in a lock context.
	for k, v := range manager.readings {
		mapCopy[k] = v
	}
	return mapCopy
}

// Read fulfills a Read request by providing the latest data read from a device
// and framing it up for the gRPC response.
func (manager *dataManager) Read(req *synse.DeviceFilter) ([]*synse.Reading, error) {
	// Validate that the incoming request has the requisite fields populated.
	err := validateDeviceFilter(req)
	if err != nil {
		log.WithField("request", req).Error("[data manager] request failed validation")
		return nil, err
	}

	// Create the id for the device.
	deviceID := makeIDString(req.Rack, req.Board, req.Device)
	err = validateForRead(deviceID)
	if err != nil {
		log.WithField("id", deviceID).Error("[data manager] unable to read device")
		return nil, err
	}

	// Get the readings for the device.
	readings := manager.getReadings(deviceID)
	if readings == nil {
		log.WithField("id", deviceID).Error("[data manager] no readings found")
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
	err := validateWriteInfo(req)
	if err != nil {
		log.WithField("request", req).Error("[data manager] request failed validation")
		return nil, err
	}

	filter := req.DeviceFilter

	// Create the id for the device.
	deviceID := makeIDString(filter.Rack, filter.Board, filter.Device)
	err = validateForWrite(deviceID)
	if err != nil {
		log.WithField("id", deviceID).Error("[data manager] unable to write to device")
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
	log.Debugf("[data manager] write response data: %#v", resp)
	return resp, nil
}
