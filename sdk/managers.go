package sdk

import (
	"sync"
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
func (w *WritingManager) Start() {
	// TODO figure out what happens here.
	logger.Info("[writing manager] started")
}
