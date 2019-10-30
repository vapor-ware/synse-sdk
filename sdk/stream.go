// Synse SDK
// Copyright (c) 2019 Vapor IO
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
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// ReadStream encapsulates a channel which is used to stream data to a client.
type ReadStream struct {
	stream   chan *ReadContext
	readings chan *ReadContext
	id       uuid.UUID
	filter   []string
	closing  bool
}

// listen collects all new readings and filters them based on the supplied filter.
func (s *ReadStream) listen() {
	log.WithFields(log.Fields{
		"filter": s.filter,
		"id":     s.id,
	}).Info("starting stream listen")
	for {
		if s.closing {
			return
		}
		r, open := <-s.stream
		if !open {
			break
		}

		if len(s.filter) == 0 {
			log.WithField("device", r.Device).Debug("collecting reading")
			if s.closing {
				return
			}
			s.readings <- r
		}

		for _, id := range s.filter {
			if r.Device == id {
				log.WithField("device", r.Device).Debug("collecting reading")
				if s.closing {
					return
				}
				s.readings <- r
				break
			}
		}
	}
}

// close the ReadStream.
func (s *ReadStream) close() {
	log.WithField("id", s.id).Info("closing read stream")
	s.closing = true
	if s.stream != nil {
		close(s.stream)
	}
	if s.readings != nil {
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
	}
}
