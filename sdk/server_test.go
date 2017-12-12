package sdk

import (
	"testing"
	"reflect"
)

func TestNewServer(t *testing.T) {
	s := NewServer(&Plugin{})
	if reflect.TypeOf(*s) != reflect.TypeOf(Server{}) {
		t.Error("NewServer did not return an instance of Server")
	}
}
