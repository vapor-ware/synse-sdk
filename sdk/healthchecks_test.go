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

//
//import (
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//)
//
//// Test_readBufferHealthCheck_Ok tests the readBufferHealthCheck when it should not error.
//func Test_readBufferHealthCheck_Ok(t *testing.T) {
//	defer func() {
//		DataManager = newDataManager()
//	}()
//
//	DataManager.readChannel = make(chan *ReadContext, 10)
//
//	err := readBufferHealthCheck()
//	assert.NoError(t, err)
//}
//
//// Test_readBufferHealthCheck_Error tests the readBufferHealthCheck when it should error.
//func Test_readBufferHealthCheck_Error(t *testing.T) {
//	defer func() {
//		DataManager = newDataManager()
//	}()
//
//	DataManager.readChannel = make(chan *ReadContext, 10)
//	for i := 0; i < 10; i++ {
//		DataManager.readChannel <- &ReadContext{}
//	}
//
//	err := readBufferHealthCheck()
//	assert.Error(t, err)
//}
//
//// Test_writeBufferHealthCheck_Ok tests the writeBufferHealthCheck when it should not error.
//func Test_writeBufferHealthCheck_Ok(t *testing.T) {
//	defer func() {
//		DataManager = newDataManager()
//	}()
//
//	DataManager.writeChannel = make(chan *WriteContext, 10)
//
//	err := writeBufferHealthCheck()
//	assert.NoError(t, err)
//}
//
//// Test_writeBufferHealthCheck_Error tests the writeBufferHealthCheck when it should error.
//func Test_writeBufferHealthCheck_Error(t *testing.T) {
//	defer func() {
//		DataManager = newDataManager()
//	}()
//
//	DataManager.writeChannel = make(chan *WriteContext, 10)
//	for i := 0; i < 10; i++ {
//		DataManager.writeChannel <- &WriteContext{}
//	}
//
//	err := writeBufferHealthCheck()
//	assert.Error(t, err)
//}
