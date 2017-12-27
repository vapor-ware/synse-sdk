package sdk

import (
	"testing"

	"github.com/vapor-ware/synse-sdk/sdk/config"
)

func TestNewDataManager(t *testing.T) {
	h := &Handlers{}
	c := config.PluginConfig{
		Name:    "test",
		Version: "test",
		Network: config.NetworkSettings{
			Type: "tcp",
			Address: "test",
		},
		Settings: config.Settings{
			Read:  config.ReadSettings{BufferSize: 200},
			Write: config.WriteSettings{BufferSize: 200},
		},
	}
	p := Plugin{handlers: h}
	err := p.SetConfig(&c)
	if err != nil {
		t.Error(err)
	}

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
	h := &Handlers{}
	c := &config.PluginConfig{
		Name:    "test",
		Version: "test",
		Network: config.NetworkSettings{
			Type: "tcp",
			Address: "test",
		},
		Settings: config.Settings{
			Read:  config.ReadSettings{BufferSize: 500},
			Write: config.WriteSettings{BufferSize: 500},
		},
	}
	p := Plugin{handlers: h}
	p.SetConfig(c)

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
