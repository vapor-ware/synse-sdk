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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/sdk/output"
	synse "github.com/vapor-ware/synse-server-grpc/go"
)

func TestNewReadContext(t *testing.T) {
	r := NewReadContext(&Device{id: "123"}, []*output.Reading{{Value: 1}})

	assert.Equal(t, "123", r.Device)
	assert.Len(t, r.Reading, 1)
}

func TestWriteData_encode(t *testing.T) {
	expected := &synse.V3WriteData{
		Data:   []byte{0, 1, 2},
		Action: "test",
	}

	wd := WriteData{
		Data:   []byte{0, 1, 2},
		Action: "test",
	}

	actual := wd.encode()

	assert.Equal(t, expected.Action, actual.Action)
	assert.Equal(t, len(expected.Data), len(actual.Data))
	for i := 0; i < len(expected.Data); i++ {
		assert.Equal(t, expected.Data[i], actual.Data[i])
	}
}

func Test_decodeWriteData(t *testing.T) {
	expected := &WriteData{
		Data:   []byte{3, 2, 1},
		Action: "test",
	}

	wd := &synse.V3WriteData{
		Data:   []byte{3, 2, 1},
		Action: "test",
	}

	actual := decodeWriteData(wd)

	assert.Equal(t, expected.Action, actual.Action)
	assert.Equal(t, len(expected.Data), len(actual.Data))
	for i := 0; i < len(expected.Data); i++ {
		assert.Equal(t, expected.Data[i], actual.Data[i])
	}
}
