package sdk

import (
	"sync"
	"time"

	"github.com/vapor-ware/synse-server-grpc/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DataManager
type DataManager struct {
	readChannel  chan *ReadContext
	writeChannel chan *WriteContext
	readings     map[string][]*Reading
	lock         *sync.Mutex
	devices      map[string]*Device // maybe this should just be a global map?
	handlers     *Handlers
}

func NewDataManager() *DataManager {
	return &DataManager{
		readChannel:  make(chan *ReadContext, Config.Settings.Read.BufferSize),
		writeChannel: make(chan *WriteContext, Config.Settings.Write.BufferSize),
		readings:     make(map[string][]*Reading),
		lock:         &sync.Mutex{},
	}
}

func (d *DataManager) registerHandlers() {

}

func (d *DataManager) goPollData() {
	go func() {
		delay := Config.Settings.LoopDelay
		for {
			d.pollWrite()
			d.pollRead()

			if delay != 0 {
				time.Sleep(time.Duration(delay) * time.Millisecond)
			}
		}
	}()
}

func (d *DataManager) pollWrite() {
	for i := 0; i < Config.Settings.Write.PerLoop; i++ {
		select {
		case w := <-d.writeChannel:
			Logger.Debugf("writing for %v (transaction: %v)", w.device, w.transaction.id)
			w.transaction.setStatusWriting()

			data := writeDataFromGRPC(w.data)
			err := d.handlers.Plugin.Write(d.devices[w.ID()], data)
			if err != nil {
				w.transaction.setStateError()
				w.transaction.message = err.Error()
				Logger.Errorf("failed to write to device %v: %v", w.device, err)
			}
			w.transaction.setStatusDone()

		default:
			break
		}
	}
}

func (d *DataManager) pollRead() {
	for _, dev := range d.devices {
		resp, err := d.handlers.Plugin.Read(dev)
		if err != nil {
			Logger.Errorf("failed to read from device %v: %v", dev.UID(), err)
		}
		d.readChannel <- resp
	}
}

func (d *DataManager) goUpdateData() {
	go func() {
		for {
			reading := <-d.readChannel
			d.lock.Lock()
			d.readings[reading.ID()] = reading.Reading
			d.lock.Unlock()
		}
	}()
}

func (d *DataManager) getReadings(device string) []*Reading {
	var r []*Reading

	d.lock.Lock()
	r = d.readings[device]
	d.lock.Unlock()
	return r
}

func (d *DataManager) Read(req *synse.ReadRequest) ([]*synse.ReadResponse, error) {
	err := validateReadRequest(req)
	if err != nil {
		return nil, err
	}

	deviceId := makeIDString(req.Rack, req.Board, req.Device)
	readings := d.getReadings(deviceId)
	if readings == nil {
		return nil, status.Errorf(
			codes.NotFound,
			"no readings found for device with id: %s", deviceId,
		)
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
