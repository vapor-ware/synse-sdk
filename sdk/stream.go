// Synse SDK
// Copyright (c) 2017-2020 Vapor IO
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package sdk

import (
	"sync"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// ReadStream encapsulates a channel which is used to stream data to a client.
type ReadStream struct {
	stream   chan *ReadContext
	readings chan *ReadContext
	id       uuid.UUID
	filter   []string
	closed   bool
	stopLock sync.Mutex
}

// listen collects all new readings and filters them based on the supplied filter.
func (s *ReadStream) listen() {
	log.WithFields(log.Fields{
		"filter": s.filter,
		"id":     s.id,
	}).Info("starting stream listen")

	defer func() {
		log.WithFields(log.Fields{
			"filter": s.filter,
			"id":     s.id,
		}).Info("terminating stream listen")
	}()

	for {
		if s.closed {
			return
		}
		r, open := <-s.stream
		if !open {
			return
		}

		if len(s.filter) == 0 {
			log.WithField("device", r.Device).Debug("collecting reading")
			s.stopLock.Lock()
			if s.closed {
				return
			}
			s.readings <- r
			s.stopLock.Unlock()
		}

		for _, id := range s.filter {
			if r.Device.id == id {
				log.WithField("device", r.Device.id).Debug("collecting reading")
				s.stopLock.Lock()
				if s.closed {
					return
				}
				s.readings <- r
				s.stopLock.Unlock()
				break
			}
		}
	}
}

// close the ReadStream.
func (s *ReadStream) close() {
	log.WithField("id", s.id).Info("closing read stream")

	s.stopLock.Lock()
	defer s.stopLock.Unlock()

	s.closed = true
	if s.stream != nil {
		// Drain the channel.
		for len(s.stream) > 0 {
			<-s.stream
		}
		// Close the channel.
		close(s.stream)
	}
	if s.readings != nil {
		// Drain the channel.
		for len(s.readings) > 0 {
			<-s.readings
		}
		// Close the channel.
		close(s.readings)
	}
}

// newReadStream creates a new ReadStream.
func newReadStream(filter []string) ReadStream {
	return ReadStream{
		stream:   make(chan *ReadContext, 128),
		readings: make(chan *ReadContext, 128),
		id:       uuid.New(),
		filter:   filter,
		stopLock: sync.Mutex{},
	}
}
