package sdk

import (
	"sync"
	"time"

	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-server-grpc/go"
)

// DataManager handles the reading from and writing to configured devices.
type DataManager struct {
	readChannel  chan *ReadContext
	writeChannel chan *WriteContext
	readings     map[string][]*Reading
	lock         *sync.Mutex
	handlers     *Handlers
	config       *config.PluginConfig
}

// NewDataManager creates a new instance of the DataManager. It initializes
// its fields appropriately, based on the current plugin configuration settings.
func NewDataManager(plugin *Plugin) *DataManager {
	return &DataManager{
		readChannel:  make(chan *ReadContext, plugin.Config.Settings.Read.BufferSize),
		writeChannel: make(chan *WriteContext, plugin.Config.Settings.Write.BufferSize),
		readings:     make(map[string][]*Reading),
		lock:         &sync.Mutex{},
		handlers:     plugin.handlers,
		config:       plugin.Config,
	}
}

// goPollData starts a go routine which acts as the read-write loop. It first
// attempts to fulfill any pending write requests, then performs reads on all
// of the configured devices.
func (d *DataManager) goPollData() {
	Logger.Debug("starting read-write poller")
	go func() {
		delay := d.config.Settings.LoopDelay
		for {
			d.pollWrite()
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
			Logger.Debugf("writing for %v (transaction: %v)", w.device, w.transaction.id)
			w.transaction.setStatusWriting()

			data := decodeWriteData(w.data)
			err := d.handlers.Plugin.Write(deviceMap[w.ID()], data)
			if err != nil {
				w.transaction.setStateError()
				w.transaction.message = err.Error()
				Logger.Errorf("failed to write to device %v: %v", w.device, err)
			}
			w.transaction.setStatusDone()

		default:
			// if there is nothing to write, do nothing
		}
	}
}

// pollRead reads from every configured device.
func (d *DataManager) pollRead() {
	for _, dev := range deviceMap {
		resp, err := d.handlers.Plugin.Read(dev)
		if err != nil {
			Logger.Errorf("failed to read from device %v: %v", dev.GUID(), err)
		}
		d.readChannel <- resp
	}
}

// goUpdateData updates the DeviceManager's readings state with the latest
// values that were read for each device.
func (d *DataManager) goUpdateData() {
	Logger.Debug("starting data updater")
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
	err := validateReadRequest(req)
	if err != nil {
		return nil, err
	}

	deviceID := makeIDString(req.Rack, req.Board, req.Device)
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
	err := validateWriteRequest(req)
	if err != nil {
		return nil, err
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
	Logger.Debugf("write response data: %#v", resp)
	return resp, nil
}
