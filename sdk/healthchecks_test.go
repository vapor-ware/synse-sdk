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
