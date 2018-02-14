package sdk

import (
	"fmt"
	"sync"
	"time"

	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
	"github.com/vapor-ware/synse-server-grpc/go"
)

// Please tell me why this file is called manager.go and not dataManger.go
// or DataManager.go. We can call it anything.

// DataManager handles the reading from and writing to configured devices.
type DataManager struct {
	readChannel  chan *ReadContext     // How to read data from a device.
	writeChannel chan *WriteContext    // How to write to a device.
	readings     map[string][]*Reading // Map of readings as strings.
	lock         *sync.Mutex           // Lock around reads and writes.
	handlers     *Handlers             // TODO: Doc
	config       *config.PluginConfig  // See config.PluginConfig.
}

// NewDataManager creates a new instance of the DataManager. It initializes
// its fields appropriately, based on the current plugin configuration settings.
// TODO: appropriately is a poor comment. Implies the reader has any idea what appropriate means.
func NewDataManager(plugin *Plugin) *DataManager {
	return &DataManager{
		// TODO: There absolutely has to be bugs in here.
		// Pointers are derefenced without checks and golang does not typically throw.
		// Solution is to check parameters and fail appropriately. (Log the error for starters.)
		readChannel:  make(chan *ReadContext, plugin.Config.Settings.Read.BufferSize),
		writeChannel: make(chan *WriteContext, plugin.Config.Settings.Write.BufferSize),
		readings:     make(map[string][]*Reading),
		lock:         &sync.Mutex{},
		handlers:     plugin.handlers,
		config:       plugin.Config,
	}
}

// writesEnabled checks to see whether writing is enableds based on the configuration.
// If the PerLoop setting is set to 0, we will never be able to write, so we consider
// writing to be disabled.
// TODO: This returns some string. Is it used outside of a log statement?
// TODO: Is the bool an error? I cannot tell in this context.
func (d *DataManager) writesEnabled() (string, bool) {
	if d.config.Settings.Write.PerLoop <= 0 {
		return "PerLoop setting <= 0", false
	}
	return "", true
}

// goPollData starts a go routine which acts as the read-write loop. It first
// attempts to fulfill any pending write requests, then performs reads on all
// of the configured devices.
func (d *DataManager) goPollData() {
	logger.Debug("starting read-write poller") // TODO: Should be logger.info.
	go func() {
		delay := d.config.Settings.LoopDelay
		for {
			if _, ok := d.writesEnabled(); ok {
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
			// Consider logger.info on writes. Writes are more important than reads.
			logger.Debugf("writing for %v (transaction: %v)", w.device, w.transaction.id)
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
	// TODO: Tell me about the device map. It's the only place deviceMap appars in this file.
	// TODO: What if deviceMap is nil?
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
	logger.Debug("starting data updater") // TODO: Should this be logger.info?
	go func() {
		for {
			reading := <-d.readChannel
			d.lock.Lock()
			d.readings[reading.ID()] = reading.Reading
			d.lock.Unlock()
		}
	}()
}

// getReadings safely gets a reading value from the DataManager readings field.
// TODO: Why does it safely get a reading? Is it because there is a lock around
// getting the reading? Would be nice to have a useful comment?
// Also - what is being locked? File lock? Thread lock? We need to know.
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
	// TODO: Comment free function. Be a team player. Tell me what this does. Tell me why it does what it does.
	err := validateReadRequest(req)
	if err != nil {
		return nil, err
	}

	deviceID := makeIDString(req.Rack, req.Board, req.Device)
	err = validateForRead(deviceID)
	if err != nil {
		return nil, err
	}

	readings := d.getReadings(deviceID)
	if readings == nil {
		return nil, notFoundErr("no readings found for device: %s", deviceID)
	}

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
	// TODO: Comment free function. Be a team player. Tell me what this does. Tell me why it does what it does.
	err := validateWriteRequest(req)
	if err != nil {
		return nil, err
	}

	deviceID := makeIDString(req.Rack, req.Board, req.Device)
	err = validateForWrite(deviceID)
	if err != nil {
		return nil, err
	}

	if ctx, enabled := d.writesEnabled(); !enabled {
		return nil, fmt.Errorf("writing is not enabled (%v)", ctx)
	}

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
