package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewServer tests that a Server is returned when the constructor
// is called.
func TestNewServer(t *testing.T) {
	s := NewServer("foo", "bar")
	assert.IsType(t, &Server{}, s)
	assert.Equal(t, "foo", s.network)
	assert.Equal(t, "bar", s.address)
}

//
//// TestSetupSocket tests setting up the socket when the socket path
//// already exists.
//func TestSetupSocket(t *testing.T) {
//	// Set up a temporary directory for testing
//	dir, err := ioutil.TempDir("", "testing")
//	assert.NoError(t, err)
//	defer func() {
//		test.CheckErr(t, os.RemoveAll(dir))
//	}()
//
//	// Set the socket path to the temp dir for the test
//	sockPath = dir
//
//	sock, err := setupSocket("test.sock")
//	if err != nil {
//		t.Error(err)
//	}
//
//	// Verify the socket path+name
//	assert.Equal(t, filepath.Join(dir, "test.sock"), sock)
//
//	// Verify that the socket path exists
//	_, err = os.Stat(sockPath)
//	assert.NoError(t, err)
//}
//
//// TestSetupSocket2 tests setting up the socket when the socket path
//// does not already exists.
//func TestSetupSocket2(t *testing.T) {
//	// Set up a temporary directory for testing
//	dir, err := ioutil.TempDir("", "testing")
//	assert.NoError(t, err)
//	// remove the temp dir now - it shouldn't exist when we set up the socket
//	test.CheckErr(t, os.RemoveAll(dir))
//
//	// Set the socket path to the temp dir for the test
//	sockPath = dir
//
//	sock, err := setupSocket("test.sock")
//	if err != nil {
//		t.Error(err)
//	}
//
//	// Verify the socket path+name
//	assert.Equal(t, filepath.Join(dir, "test.sock"), sock)
//
//	// Verify that the socket path exists
//	_, err = os.Stat(sockPath)
//	assert.NoError(t, err)
//}
//
//// TestSetupSocket2 tests setting up the socket when the socket path
//// and the socket itself already exist.
//func TestSetupSocket3(t *testing.T) {
//	// Set up a temporary directory for testing
//	dir, err := ioutil.TempDir("", "testing")
//	assert.NoError(t, err)
//	defer func() {
//		test.CheckErr(t, os.RemoveAll(dir))
//	}()
//
//	// Set the socket path to the temp dir for the test
//	sockPath = dir
//
//	// Make the socket file
//	filename := filepath.Join(dir, "test.sock")
//	_, err = os.Create(filename)
//	assert.NoError(t, err)
//
//	sock, err := setupSocket("test.sock")
//	if err != nil {
//		t.Error(err)
//	}
//
//	// Verify the socket path+name
//	assert.Equal(t, filepath.Join(dir, "test.sock"), sock)
//
//	// Verify that the socket path exists
//	_, err = os.Stat(sockPath)
//	assert.NoError(t, err)
//
//	// Verify that the socket itself no longer exists (setupSocket cleans
//	// up old socket instances)
//	_, err = os.Stat(filename)
//	exists := !os.IsNotExist(err)
//	assert.False(t, exists, "socket should no longer exist, but still does")
//}
