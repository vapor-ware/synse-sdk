package sdk

import (
	"sync"
	"github.com/vapor-ware/synse-server-grpc/go"
)


//
type ReadingManager struct {
	channel chan ReadResource
	values map[string][]Reading

	lock *sync.Mutex
}

//
func (r *ReadingManager) Start() {

	r.lock = &sync.Mutex{}

	go func() {
		for {
			reading := <- r.channel
			r.lock.Lock()
			r.values[reading.Device] = reading.Reading
			r.lock.Unlock()
		}
	}()
	logger.Info("[reading manager] started")
}

//
func (r *ReadingManager) GetReading(key string) []Reading {
	var reading []Reading
	r.lock.Lock()
	reading = r.values[key]
	r.lock.Unlock()
	return reading
}


//
type WritingManager struct {
	channel chan WriteResource
	values map[string]string
}


//
func (w *WritingManager) Write(in *synse.WriteRequest) map[string]*synse.WriteData {

	var response = make(map[string]*synse.WriteData)

	for _, data := range in.Data {
		transaction := NewTransactionId()
		UpdateTransactionStatus(transaction.id, PENDING)
		response[transaction.id] = data

		w.channel <- WriteResource{
			transaction,
			in.Uid,
			data,
		}
	}
	logger.Debugf("Write response data: %+v", response)
	return response
}
