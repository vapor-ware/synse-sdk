package sdk

import (
	"fmt"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/vapor-ware/synse-server-grpc/go"
)

// ReadingManager is used to manage the reading of devices in an asynchronous
// manner. It contains a channel used to get the readings from the read-write
// loop and update internal state to reflect those reading values.
type ReadingManager struct {
	channel chan *ReadResource
	values  map[string][]*Reading
	lock    *sync.Mutex
}

// start runs a goroutine that reads from the ReadResource channel and
// updates the values map with the readings it gets from the channel.
func (r *ReadingManager) start() {

	r.lock = &sync.Mutex{}

	go func() {
		for {
			reading := <- r.channel
			r.lock.Lock()
			r.values[reading.Device] = reading.Reading
			r.lock.Unlock()
		}
	}()
	Logger.Info("[reading manager] started")
}

// read is used to get all readings associated with the specified device. If
// the specified device is not found in the reading manager's state, no
// readings will be returned.
func (r *ReadingManager) read(in *synse.ReadRequest) ([]*synse.ReadResponse, error) {
	uid := in.GetUid()
	if uid == "" {
		return nil, status.Errorf(codes.InvalidArgument, "No UID supplied to Read.")
	}
	Logger.Debugf("Reading uid: %v", uid)


	readings := r.getReadings(uid)
	if readings == nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("No readings found for device %v", uid))
	}

	var response []*synse.ReadResponse
	for _, r := range readings {
		reading := &synse.ReadResponse{
			Timestamp: r.Timestamp,
			Type: r.Type,
			Value: r.Value,
		}
		response = append(response, reading)
	}
	return response, nil
}

// getReadings gets the reading(s) for the the specified device from the
// values map.
func (r *ReadingManager) getReadings(key string) []*Reading {
	var reading []*Reading
	r.lock.Lock()
	reading = r.values[key]
	r.lock.Unlock()
	return reading
}


// WritingManager is used to manage the writing of data to devices. It has a
// channel that is used to push pending writes to. This channel acts as a job
// queue which is worked on at the top of the read-write loop.
type WritingManager struct {
	channel chan *WriteResource
}


// write is used to put new data from WriteRequests onto the write queue.
// Each WriteRequest contains a slice of WriteData. Each of these WriteData
// are put on the queue as their own jobs and are given individual transaction
// ids. The transaction ids are mapped back to the WriteData which is returned.
// The WriteData is returned with the transaction ids to help provide context
// for differentiating transaction ids in the event that a WriteRequest has
// more than one WriteData.
func (w *WritingManager) write(in *synse.WriteRequest) map[string]*synse.WriteData {

	var response = make(map[string]*synse.WriteData)

	for _, data := range in.Data {
		transaction := NewTransactionID()
		UpdateTransactionStatus(transaction.id, PENDING)
		response[transaction.id] = data

		w.channel <- &WriteResource{
			transaction,
			in.Uid,
			data,
		}
	}
	Logger.Debugf("Write response data: %+v", response)
	return response
}
