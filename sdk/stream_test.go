// Synse SDK
// Copyright (c) 2019-2020 Vapor IO
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
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/sdk/output"
)

func TestNewReadStream(t *testing.T) {
	s := newReadStream([]string{"foo", "bar"})

	assert.NotNil(t, s.stream)
	assert.NotNil(t, s.readings)
	assert.NotNil(t, s.id)
	assert.Equal(t, []string{"foo", "bar"}, s.filter)
	assert.False(t, s.closing)
}

func TestReadStream_close_openChannels(t *testing.T) {
	streamChan := make(chan *ReadContext, 3)
	readingsChan := make(chan *ReadContext, 3)

	s := ReadStream{
		stream:   streamChan,
		readings: readingsChan,
	}

	streamChan <- &ReadContext{}
	readingsChan <- &ReadContext{}

	_, open := <-streamChan
	assert.True(t, open)
	_, open = <-readingsChan
	assert.True(t, open)
	assert.False(t, s.closing)

	s.close()

	assert.True(t, s.closing)
	_, open = <-streamChan
	assert.False(t, open)
	_, open = <-readingsChan
	assert.False(t, open)
}

func TestReadStream_close_nilChannels(t *testing.T) {
	s := ReadStream{}
	assert.Nil(t, s.stream)
	assert.Nil(t, s.readings)
	assert.False(t, s.closing)

	assert.NotPanics(t, func() {
		s.close()
	})
	assert.True(t, s.closing)
}

func TestReadStream_listen_withFilter(t *testing.T) {
	s := ReadStream{
		stream:   make(chan *ReadContext, 128),
		readings: make(chan *ReadContext, 128),
		id:       uuid.New(),
		filter:   []string{"12345", "11111"},
	}
	assert.False(t, s.closing)

	// Create read contexts to send to the stream.
	ctxs := []*ReadContext{
		{
			Device:  "11111",
			Reading: []*output.Reading{{Value: 1}},
		},
		{
			Device:  "22222",
			Reading: []*output.Reading{{Value: 1}},
		},
		{
			Device:  "12345",
			Reading: []*output.Reading{{Value: 1}},
		},
		{
			Device:  "54321",
			Reading: []*output.Reading{{Value: 1}},
		},
	}

	for _, c := range ctxs {
		s.stream <- c
	}
	// Be sure to close the stream so we break out of the listen loop
	close(s.stream)

	s.listen()

	// Collect the readings from the stream
	readings := []*ReadContext{}
	var done bool
	for {
		select {
		case r := <-s.readings:
			readings = append(readings, r)
		case <-time.After(600 * time.Millisecond):
			done = true
			break
		}
		if done {
			break
		}
	}

	// Ensure that only the devices specified by the match filter were collected.
	assert.Len(t, readings, 2)
	assert.Equal(t, "11111", readings[0].Device)
	assert.Equal(t, "12345", readings[1].Device)

	// We did not call close(), so the stream should not be in the closing state
	assert.False(t, s.closing)
}

func TestReadStream_listen_noFilter(t *testing.T) {
	s := ReadStream{
		stream:   make(chan *ReadContext, 128),
		readings: make(chan *ReadContext, 128),
		id:       uuid.New(),
		filter:   []string{},
	}
	assert.False(t, s.closing)

	// Create read contexts to send to the stream.
	ctxs := []*ReadContext{
		{
			Device:  "11111",
			Reading: []*output.Reading{{Value: 1}},
		},
		{
			Device:  "22222",
			Reading: []*output.Reading{{Value: 1}},
		},
		{
			Device:  "12345",
			Reading: []*output.Reading{{Value: 1}},
		},
		{
			Device:  "54321",
			Reading: []*output.Reading{{Value: 1}},
		},
	}

	for _, c := range ctxs {
		s.stream <- c
	}
	// Be sure to close the stream so we break out of the listen loop
	close(s.stream)

	s.listen()

	// Collect the readings from the stream
	readings := []*ReadContext{}
	var done bool
	for {
		select {
		case r := <-s.readings:
			readings = append(readings, r)
		case <-time.After(600 * time.Millisecond):
			done = true
			break
		}
		if done {
			break
		}
	}

	// Ensure that all devices were collected since there is no filter.
	assert.Len(t, readings, 4)
	assert.Equal(t, "11111", readings[0].Device)
	assert.Equal(t, "22222", readings[1].Device)
	assert.Equal(t, "12345", readings[2].Device)
	assert.Equal(t, "54321", readings[3].Device)

	// We did not call close(), so the stream should not be in the closing state
	assert.False(t, s.closing)
}

func TestReadStream_listen_close(t *testing.T) {
	s := ReadStream{
		stream:   make(chan *ReadContext, 128),
		readings: make(chan *ReadContext, 128),
		id:       uuid.New(),
		filter:   []string{},
	}
	assert.False(t, s.closing)

	// Create read contexts to send to the stream.
	ctxs := []*ReadContext{
		{
			Device:  "11111",
			Reading: []*output.Reading{{Value: 1}},
		},
		{
			Device:  "22222",
			Reading: []*output.Reading{{Value: 1}},
		},
		{
			Device:  "12345",
			Reading: []*output.Reading{{Value: 1}},
		},
		{
			Device:  "54321",
			Reading: []*output.Reading{{Value: 1}},
		},
	}

	for _, c := range ctxs {
		s.stream <- c
	}
	s.close()
	s.listen()

	_, open := <-s.readings
	assert.False(t, open)

	assert.True(t, s.closing)
}
