package sdk

import (
	"testing"
)

func TestNewDataManager(t *testing.T) {
	Config.Settings.Read.BufferSize = 200
	Config.Settings.Write.BufferSize = 200

	h := &Handlers{}
	p := Plugin{handlers: h}

	d := NewDataManager(&p)

	if cap(d.writeChannel) != 200 {
		t.Errorf("write channel should be of size 200 but is %v", len(d.writeChannel))
	}
	if cap(d.readChannel) != 200 {
		t.Errorf("read channel should be of size 200 but is %v", len(d.readChannel))
	}
	if d.handlers != h {
		t.Error("handler is not the expected handler instance")
	}
}

func TestNewDataManager2(t *testing.T) {
	Config.Settings.Read.BufferSize = 500
	Config.Settings.Write.BufferSize = 500

	h := &Handlers{}
	p := Plugin{handlers: h}

	d := NewDataManager(&p)

	if cap(d.writeChannel) != 500 {
		t.Errorf("write channel should be of size 500 but is %v", len(d.writeChannel))
	}
	if cap(d.readChannel) != 500 {
		t.Errorf("read channel should be of size 500 but is %v", len(d.readChannel))
	}
	if d.handlers != h {
		t.Error("handler is not the expected handler instance")
	}
}
