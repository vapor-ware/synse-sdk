package sdk

import (
	"fmt"
	"sync"
	"time"

	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
	"github.com/vapor-ware/synse-server-grpc/go"
)

// TODO: Rename file prior to release: https://github.com/vapor-ware/synse-sdk/issues/119

// DataManager handles the reading from and writing to configured devices.
type DataManager struct {
	readChannel  chan *ReadContext     // Channel to get data from the goroutine that reads from devices.
	writeChannel chan *WriteContext    // Channel to pass data to the goroutine that writes to devices.
	readings     map[string][]*Reading // Map of readings as strings. Key is the device UID.
	lock         *sync.Mutex           // Lock around asynch reads and writes.
	handlers     *Handlers             // See sdk/handlers.go.
	config       *config.PluginConfig  // See config.PluginConfig.
}

// NewDataManager creates a new instance of the DataManager using the existing
// configurations and handlers registered with the plugin.
func NewDataManager(plugin *Plugin) *DataManager {
	return &DataManager{
		// TODO: https://github.com/vapor-ware/synse-sdk/issues/118
		readChannel:  make(chan *ReadContext, plugin.Config.Settings.Read.BufferSize),
		writeChannel: make(chan *WriteContext, plugin.Config.Settings.Write.BufferSize),
		readings:     make(map[string][]*Reading),
		lock:         &sync.Mutex{},
		handlers:     plugin.handlers,
		config:       plugin.Config,
	}
}

// writesEnabled checks to see whether writing is enabled based on the configuration.
// If the PerLoop setting is <= 0, we will never be able to write, so we consider
// writing to be disabled.
func (d *DataManager) writesEnabled() bool {
	return d.config.Settings.Write.PerLoop > 0
}

// goPollData starts a go routine which acts as the read-write loop. It first
// attempts to fulfill any pending write requests, then performs reads on all
// of the configured devices.
func (d *DataManager) goPollData() {
	logger.Info("starting read-write poller")
	go func() {
		delay := d.config.Settings.LoopDelay
		for {
			if ok := d.writesEnabled(); ok {
				d.pollWrite()
			}
			d.pollRead()

			if delay != 0 {
				time.Sleep(time.Duration(delay) * time.Millisecond)
			}
		}
	}()
}

// pollWrite checks for any pending writes and, if any exist, attempts to fulfill
// the writes and update the transaction state accordingly.
func (d *DataManager) pollWrite() {
	for i := 0; i < d.config.Settings.Write.PerLoop; i++ {
		select {
		case w := <-d.writeChannel:
			// If this is too chatty we can change back to logger.Debugf.
			logger.Infof("writing for %v (transaction: %v)", w.device, w.transaction.id)
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

		default:
			// if there is nothing to write, do nothing
		}
	}
}

// pollRead reads from every configured device.
func (d *DataManager) pollRead() {
	// deviceMap is a non-nil global in sdk/devices.go containing a single Device
	// struct instance per configured device.
	for _, dev := range deviceMap {
		resp, err := dev.Read()
		if err != nil {
			logger.Errorf("failed to read from device %v: %v", dev.GUID(), err)
		} else {
			d.readChannel <- resp
		}
	}
}

// goUpdateData updates the DeviceManager's readings state with the latest
// values that were read for each device.
func (d *DataManager) goUpdateData() {
	logger.Info("starting data updater")
	go func() {
		for {
			reading := <-d.readChannel
			d.lock.Lock()
			d.readings[reading.ID()] = reading.Reading
			d.lock.Unlock()
		}
	}()
}

// getReadings safely gets a reading value from the DataManager readings field by
// accessing the readings for the specified device within a lock context. Since the
// readings map is updated in a separate goroutine, we want to lock access around the
// map to prevent simultaneous access collisions.
func (d *DataManager) getReadings(device string) []*Reading {
	var r []*Reading

	d.lock.Lock()
	r = d.readings[device]
	d.lock.Unlock()
	return r
}

// Read fulfills a Read request by providing the latest data read from a device
// and framing it up for the gRPC response.
func (d *DataManager) Read(req *synse.ReadRequest) ([]*synse.ReadResponse, error) {
	// Parameter check.
	err := validateReadRequest(req)
	if err != nil {
		return nil, err
	}

	// Create the id for the device.
	deviceID := makeIDString(req.Rack, req.Board, req.Device)
	err = validateForRead(deviceID)
	if err != nil {
		return nil, err
	}

	// Get the readings for the device.
	readings := d.getReadings(deviceID)
	if readings == nil {
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

// Write fulfills a Write request by queuing up the write transaction and framing
// up the corresponding gRPC response.
func (d *DataManager) Write(req *synse.WriteRequest) (map[string]*synse.WriteData, error) {
	// Parameter check.
	err := validateWriteRequest(req)
	if err != nil {
		return nil, err
	}

	// Create the id for the device.
	deviceID := makeIDString(req.Rack, req.Board, req.Device)
	err = validateForWrite(deviceID)
	if err != nil {
		return nil, err
	}

	// Ensure writes are enabled.
	if enabled := d.writesEnabled(); !enabled {
		return nil, fmt.Errorf("writing is not enabled")
	}

	// Perform the write and build the response.
	var resp = make(map[string]*synse.WriteData)
	for _, data := range req.Data {
		t := NewTransaction()
		t.setStatusPending()

		resp[t.id] = data
		d.writeChannel <- &WriteContext{
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