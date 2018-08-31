package sdk

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/internal/test"
	"github.com/vapor-ware/synse-sdk/sdk/health"
	"github.com/vapor-ware/synse-server-grpc/go"
)

// TestNewServer tests that a server is returned when the constructor
// is called.
func TestNewServer(t *testing.T) {
	s := newServer("foo", "bar")
	assert.IsType(t, &server{}, s)
	assert.Equal(t, "foo", s.network)
	assert.Equal(t, "bar", s.address)
}

// TestServer_setup_TCP tests successfully setting up the server in TCP mode.
func TestServer_setup_TCP(t *testing.T) {
	defer resetContext()

	s := newServer(networkTypeTCP, "test")
	err := s.setup()
	assert.NoError(t, err)
}

// TestServer_setup_Unix tests successfully setting up the server in Unix mode.
func TestServer_setup_Unix(t *testing.T) {
	defer resetContext()

	// Set up a temporary directory for test data.
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	sockPath = test.TempDir

	// first, check that there are no post-run actions
	assert.Equal(t, 0, len(ctx.postRunActions))

	s := newServer(networkTypeUnix, "test")
	err := s.setup()
	assert.NoError(t, err)

	// now, there should be one post run action
	assert.Equal(t, 1, len(ctx.postRunActions))
}

// TestServer_setup_Unknown tests successfully setting up the server in an unknown mode.
func TestServer_setup_Unknown(t *testing.T) {
	defer resetContext()

	s := newServer("foo", "test")
	err := s.setup()
	assert.Error(t, err)
}

// TestServer_cleanup_TCP tests cleaning up the server in TCP mode.
func TestServer_cleanup_TCP(t *testing.T) {
	s := newServer(networkTypeTCP, "test")
	err := s.cleanup()
	assert.NoError(t, err)
}

// TestServer_cleanup_Unix tests cleaning up the server in Unix mode.
func TestServer_cleanup_Unix(t *testing.T) {
	s := newServer(networkTypeUnix, "test")
	err := s.cleanup()
	assert.NoError(t, err)
}

// TestServer_cleanup_Unknown tests cleaning up the server in an unknown mode.
func TestServer_cleanup_Unknown(t *testing.T) {
	s := newServer("foo", "test")
	err := s.cleanup()
	assert.Error(t, err)
}

// TestServer_Serve tests serving when there is a setup error.
func TestServer_Serve(t *testing.T) {
	defer resetContext()
	defer Config.reset()
	Config.Plugin = &PluginConfig{
		Network: &NetworkSettings{},
	}

	s := newServer("foo", "bar")
	defer s.Stop()
	err := s.Serve()
	assert.Error(t, err)
}

// TestServer_Serve2 tests serving when there is an error creating a net listener.
func TestServer_Serve2(t *testing.T) {
	defer resetContext()
	defer Config.reset()
	Config.Plugin = &PluginConfig{
		Network: &NetworkSettings{},
	}

	s := newServer("tcp", "bar")
	defer s.Stop()
	err := s.Serve()
	assert.Error(t, err)
}

// TestServer_Test tests the Test method of the gRPC plugin service.
func TestServer_Test(t *testing.T) {
	s := server{}
	req := &synse.Empty{}
	resp, err := s.Test(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, true, resp.Ok)
}

// TestServer_Version tests the Version method of the gRPC plugin service.
func TestServer_Version(t *testing.T) {
	s := server{}
	req := &synse.Empty{}
	resp, err := s.Version(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, version.Arch, resp.Arch)
	assert.Equal(t, version.OS, resp.Os)
	assert.Equal(t, version.SDKVersion, resp.SdkVersion)
	assert.Equal(t, version.BuildDate, resp.BuildDate)
	assert.Equal(t, version.GitCommit, resp.GitCommit)
	assert.Equal(t, version.GitTag, resp.GitTag)
	assert.Equal(t, version.PluginVersion, resp.PluginVersion)
}

// TestServer_Health tests the Health method of the gRPC plugin service when
// there are no health checks defined.
func TestServer_Health(t *testing.T) {
	s := server{}
	req := &synse.Empty{}
	resp, err := s.Health(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Timestamp)
	assert.Equal(t, synse.PluginHealth_OK, resp.Status)
	assert.Equal(t, 0, len(resp.Checks))
}

// TestServer_Health2 tests the Health method of the gRPC plugin service when
// there is a passing health check.
func TestServer_Health2(t *testing.T) {
	defer func() {
		health.DefaultCatalog = health.NewCatalog()
	}()

	health.Register("foo", health.NewChecker("foo"))

	s := server{}
	req := &synse.Empty{}
	resp, err := s.Health(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Timestamp)
	assert.Equal(t, synse.PluginHealth_OK, resp.Status)
	assert.Equal(t, 1, len(resp.Checks))
}

// TestServer_Health3 tests the Health method of the gRPC plugin service when
// there is a failing health check.
func TestServer_Health3(t *testing.T) {
	defer func() {
		health.DefaultCatalog = health.NewCatalog()
	}()

	checker := health.NewChecker("foo")
	checker.Update(fmt.Errorf("err"))
	health.Register("foo", checker)

	s := server{}
	req := &synse.Empty{}
	resp, err := s.Health(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Timestamp)
	assert.Equal(t, synse.PluginHealth_FAILING, resp.Status)
	assert.Equal(t, 1, len(resp.Checks))
}

// TestServer_Health4 tests the Health method of the gRPC plugin service when
// there is a failing health check and a passing health check.
func TestServer_Health4(t *testing.T) {
	defer func() {
		health.DefaultCatalog = health.NewCatalog()
	}()

	health.Register("foo", health.NewChecker("foo"))

	checker := health.NewChecker("foo")
	checker.Update(fmt.Errorf("err"))
	health.Register("bar", checker)

	s := server{}
	req := &synse.Empty{}
	resp, err := s.Health(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Timestamp)
	assert.Equal(t, synse.PluginHealth_PARTIALLY_DEGRADED, resp.Status)
	assert.Equal(t, 2, len(resp.Checks))
}

// TestServer_Capabilities tests the Capabilities method of the gRPC plugin service.
func TestServer_Capabilities(t *testing.T) {
	s := server{}
	req := &synse.Empty{}
	mock := test.NewMockCapabilitiesStream()
	err := s.Capabilities(req, mock)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(mock.Results))
}

// TestServer_Capabilities2 tests the Capabilities method of the gRPC plugin service
// when there are actual devices to get capabilities from.
func TestServer_Capabilities2(t *testing.T) {
	defer resetContext()

	ctx.devices["foo"] = &Device{
		Kind: "foo",
		Outputs: []*Output{
			{OutputType: OutputType{Name: "output1"}},
			{OutputType: OutputType{Name: "output2"}},
		},
	}
	ctx.devices["bar"] = &Device{
		Kind: "bar",
		Outputs: []*Output{
			{OutputType: OutputType{Name: "output3"}},
		},
	}

	s := server{}
	req := &synse.Empty{}
	mock := test.NewMockCapabilitiesStream()
	err := s.Capabilities(req, mock)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(mock.Results))
	assert.Equal(t, 2, len(mock.Results["foo"].GetOutputs()))
	assert.Equal(t, 1, len(mock.Results["bar"].GetOutputs()))
}

// TestServer_Capabilities3 tests the Capabilities method of the gRPC plugin service
// when there is an error returned.
func TestServer_Capabilities3(t *testing.T) {
	defer resetContext()

	ctx.devices["foo"] = &Device{
		Kind: "foo",
		Outputs: []*Output{
			{OutputType: OutputType{Name: "output1"}},
			{OutputType: OutputType{Name: "output2"}},
		},
	}
	ctx.devices["bar"] = &Device{
		Kind: "bar",
		Outputs: []*Output{
			{OutputType: OutputType{Name: "output3"}},
		},
	}

	s := server{}
	req := &synse.Empty{}
	mock := &test.MockCapabilitiesStreamErr{}
	err := s.Capabilities(req, mock)

	assert.Error(t, err)
}

// TestServer_Devices tests the Devices method of the gRPC plugin service.
func TestServer_Devices(t *testing.T) {
	s := server{}
	req := &synse.DeviceFilter{}
	mock := test.NewMockDevicesStream()
	err := s.Devices(req, mock)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(mock.Results))
}

// TestServer_Devices_NoFilter tests the Devices method of the gRPC plugin service when
// there are devices to get.
func TestServer_Devices_NoFilter(t *testing.T) {
	defer resetContext()

	foo := &Device{
		Kind: "foo",
		Location: &Location{
			Rack:  "rack-1",
			Board: "board-1",
		},
		Outputs: []*Output{
			{OutputType: OutputType{Name: "output1"}},
			{OutputType: OutputType{Name: "output2"}},
		},
	}
	bar := &Device{
		Kind: "bar",
		Location: &Location{
			Rack:  "rack-1",
			Board: "board-2",
		},
		Outputs: []*Output{
			{OutputType: OutputType{Name: "output3"}},
		},
	}
	baz := &Device{
		Kind: "baz",
		Location: &Location{
			Rack:  "rack-2",
			Board: "board-1",
		},
		Outputs: []*Output{
			{OutputType: OutputType{Name: "output4"}},
		},
	}

	ctx.devices["foo"] = foo
	ctx.devices["bar"] = bar
	ctx.devices["baz"] = baz

	s := server{}
	req := &synse.DeviceFilter{}
	mock := test.NewMockDevicesStream()
	err := s.Devices(req, mock)

	assert.NoError(t, err)
	assert.Equal(t, 3, len(mock.Results))
	assert.NotNil(t, mock.Results[foo.ID()])
	assert.NotNil(t, mock.Results[bar.ID()])
	assert.NotNil(t, mock.Results[baz.ID()])
}

// TestServer_Devices_FilterRack tests the Devices method of the gRPC plugin service when
// there are devices to get, and we filter on rack.
func TestServer_Devices_FilterRack(t *testing.T) {
	defer resetContext()

	foo := &Device{
		Kind: "foo",
		Location: &Location{
			Rack:  "rack-1",
			Board: "board-1",
		},
		Outputs: []*Output{
			{OutputType: OutputType{Name: "output1"}},
			{OutputType: OutputType{Name: "output2"}},
		},
	}
	bar := &Device{
		Kind: "bar",
		Location: &Location{
			Rack:  "rack-1",
			Board: "board-2",
		},
		Outputs: []*Output{
			{OutputType: OutputType{Name: "output3"}},
		},
	}
	baz := &Device{
		Kind: "baz",
		Location: &Location{
			Rack:  "rack-2",
			Board: "board-1",
		},
		Outputs: []*Output{
			{OutputType: OutputType{Name: "output4"}},
		},
	}

	ctx.devices["foo"] = foo
	ctx.devices["bar"] = bar
	ctx.devices["baz"] = baz

	s := server{}
	req := &synse.DeviceFilter{
		Rack: "rack-1",
	}
	mock := test.NewMockDevicesStream()
	err := s.Devices(req, mock)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(mock.Results))
	assert.NotNil(t, mock.Results[foo.ID()])
	assert.NotNil(t, mock.Results[bar.ID()])
	assert.Nil(t, mock.Results[baz.ID()])
}

// TestServer_Devices_FilterBoard tests the Devices method of the gRPC plugin service when
// there are devices to get, and we filter on board.
func TestServer_Devices_FilterBoard(t *testing.T) {
	defer resetContext()

	foo := &Device{
		Kind: "foo",
		Location: &Location{
			Rack:  "rack-1",
			Board: "board-1",
		},
		Outputs: []*Output{
			{OutputType: OutputType{Name: "output1"}},
			{OutputType: OutputType{Name: "output2"}},
		},
	}
	bar := &Device{
		Kind: "bar",
		Location: &Location{
			Rack:  "rack-1",
			Board: "board-2",
		},
		Outputs: []*Output{
			{OutputType: OutputType{Name: "output3"}},
		},
	}
	baz := &Device{
		Kind: "baz",
		Location: &Location{
			Rack:  "rack-2",
			Board: "board-1",
		},
		Outputs: []*Output{
			{OutputType: OutputType{Name: "output4"}},
		},
	}

	ctx.devices["foo"] = foo
	ctx.devices["bar"] = bar
	ctx.devices["baz"] = baz

	s := server{}
	req := &synse.DeviceFilter{
		Rack:  "rack-1",
		Board: "board-1",
	}
	mock := test.NewMockDevicesStream()
	err := s.Devices(req, mock)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(mock.Results))
	assert.NotNil(t, mock.Results[foo.ID()])
	assert.Nil(t, mock.Results[bar.ID()])
	assert.Nil(t, mock.Results[baz.ID()])
}

// TestServer_Devices_FilterNoMatch tests the Devices method of the gRPC plugin service when
// there are devices to get, but the filter does not match any of them.
func TestServer_Devices_FilterNoMatch(t *testing.T) {
	defer resetContext()

	foo := &Device{
		Kind: "foo",
		Location: &Location{
			Rack:  "rack-1",
			Board: "board-1",
		},
		Outputs: []*Output{
			{OutputType: OutputType{Name: "output1"}},
			{OutputType: OutputType{Name: "output2"}},
		},
	}
	bar := &Device{
		Kind: "bar",
		Location: &Location{
			Rack:  "rack-1",
			Board: "board-2",
		},
		Outputs: []*Output{
			{OutputType: OutputType{Name: "output3"}},
		},
	}
	baz := &Device{
		Kind: "baz",
		Location: &Location{
			Rack:  "rack-2",
			Board: "board-1",
		},
		Outputs: []*Output{
			{OutputType: OutputType{Name: "output4"}},
		},
	}

	ctx.devices["foo"] = foo
	ctx.devices["bar"] = bar
	ctx.devices["baz"] = baz

	s := server{}
	req := &synse.DeviceFilter{
		Rack: "rack-3",
	}
	mock := test.NewMockDevicesStream()
	err := s.Devices(req, mock)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(mock.Results))
	assert.Nil(t, mock.Results[foo.ID()])
	assert.Nil(t, mock.Results[bar.ID()])
	assert.Nil(t, mock.Results[baz.ID()])
}

// TestServer_Devices_FilterDevice tests the Devices method of the gRPC plugin service when
// there are devices to get, and we filter on device. We disallow filtering on device.
func TestServer_Devices_FilterDevice(t *testing.T) {
	s := server{}
	req := &synse.DeviceFilter{
		Rack:   "rack-1",
		Board:  "board-1",
		Device: "device-1",
	}
	mock := test.NewMockDevicesStream()
	err := s.Devices(req, mock)

	assert.Error(t, err)
	assert.Equal(t, 0, len(mock.Results))
}

// TestServer_Devices_Error tests the Devices method of the gRPC plugin service when
// an error is returned because a device is specified.
func TestServer_Devices_Error(t *testing.T) {
	defer resetContext()

	foo := &Device{
		Kind: "foo",
		Location: &Location{
			Rack:  "rack-1",
			Board: "board-1",
		},
		Outputs: []*Output{
			{OutputType: OutputType{Name: "output1"}},
			{OutputType: OutputType{Name: "output2"}},
		},
	}
	bar := &Device{
		Kind: "bar",
		Location: &Location{
			Rack:  "rack-1",
			Board: "board-2",
		},
		Outputs: []*Output{
			{OutputType: OutputType{Name: "output3"}},
		},
	}
	baz := &Device{
		Kind: "baz",
		Location: &Location{
			Rack:  "rack-2",
			Board: "board-1",
		},
		Outputs: []*Output{
			{OutputType: OutputType{Name: "output4"}},
		},
	}

	ctx.devices["foo"] = foo
	ctx.devices["bar"] = bar
	ctx.devices["baz"] = baz

	s := server{}
	req := &synse.DeviceFilter{
		Rack:  "rack-1",
		Board: "board-1",
	}
	mock := &test.MockDevicesStreamErr{}
	err := s.Devices(req, mock)

	assert.Error(t, err)
}

// TestServer_Devices_Error2 tests the Devices method of the gRPC plugin service when
// an error is returned because a board is specified with no rack.
func TestServer_Devices_Error2(t *testing.T) {
	s := server{}
	req := &synse.DeviceFilter{
		Board: "board",
	}
	mock := &test.MockDevicesStreamErr{}
	err := s.Devices(req, mock)

	assert.Error(t, err)
}

// TestServer_Metainfo tests the Metainfo method of the gRPC plugin service.
func TestServer_Metainfo(t *testing.T) {
	s := server{}
	req := &synse.Empty{}
	resp, err := s.Metainfo(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, metainfo.Name, resp.GetName())
	assert.Equal(t, metainfo.Maintainer, resp.GetMaintainer())
	assert.Equal(t, metainfo.Description, resp.GetDescription())
	assert.Equal(t, metainfo.VCS, resp.GetVcs())

	v := resp.GetVersion()
	assert.Equal(t, version.Arch, v.Arch)
	assert.Equal(t, version.OS, v.Os)
	assert.Equal(t, version.SDKVersion, v.SdkVersion)
	assert.Equal(t, version.BuildDate, v.BuildDate)
	assert.Equal(t, version.GitCommit, v.GitCommit)
	assert.Equal(t, version.GitTag, v.GitTag)
	assert.Equal(t, version.PluginVersion, v.PluginVersion)
}

// TestServer_Read tests the Read method of the gRPC plugin service.
func TestServer_Read(t *testing.T) {
	defer func() {
		DataManager = newDataManager()
		resetContext()
		Config.reset()
	}()

	Config.Plugin = &PluginConfig{
		Settings: &PluginSettings{
			Read: &ReadSettings{
				Enabled: true,
			},
		},
	}
	ctx.devices["rack-board-device"] = &Device{
		id:   "device",
		Kind: "foo",
		Location: &Location{
			Rack:  "rack",
			Board: "board",
		},
		Outputs: []*Output{
			{OutputType: OutputType{Name: "output1"}},
			{OutputType: OutputType{Name: "output2"}},
		},
		Handler: &DeviceHandler{
			Read: func(device *Device) ([]*Reading, error) {
				return nil, nil
			},
		},
	}
	DataManager.readings["rack-board-device"] = []*Reading{
		{
			Timestamp: "now",
			Type:      "temperature",
			Value:     3,
		},
		{
			Timestamp: "now",
			Type:      "humidity",
			Value:     5,
		},
	}

	s := server{}
	req := &synse.DeviceFilter{
		Rack:   "rack",
		Board:  "board",
		Device: "device",
	}
	mock := test.NewMockReadStream()
	err := s.Read(req, mock)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(mock.Results))
}

// TestServer_Read2 tests the Read method of the gRPC plugin service when
// the filter does not match anything.
func TestServer_Read2(t *testing.T) {
	s := server{}
	req := &synse.DeviceFilter{
		Rack:   "rack",
		Board:  "board",
		Device: "device",
	}
	mock := test.NewMockReadStream()
	err := s.Read(req, mock)

	assert.Error(t, err)
	assert.Equal(t, 0, len(mock.Results))
}

// TestServer_Read3 tests the Read method of the gRPC plugin service when
// sending to the stream results in error.
func TestServer_Read3(t *testing.T) {
	defer func() {
		DataManager = newDataManager()
		resetContext()
		Config.reset()
	}()

	Config.Plugin = &PluginConfig{
		Settings: &PluginSettings{
			Read: &ReadSettings{
				Enabled: true,
			},
		},
	}
	ctx.devices["rack-board-device"] = &Device{
		id:   "device",
		Kind: "foo",
		Location: &Location{
			Rack:  "rack",
			Board: "board",
		},
		Outputs: []*Output{
			{OutputType: OutputType{Name: "output1"}},
			{OutputType: OutputType{Name: "output2"}},
		},
		Handler: &DeviceHandler{
			Read: func(device *Device) ([]*Reading, error) {
				return nil, nil
			},
		},
	}
	DataManager.readings["rack-board-device"] = []*Reading{
		{
			Timestamp: "now",
			Type:      "temperature",
			Value:     3,
		},
		{
			Timestamp: "now",
			Type:      "humidity",
			Value:     5,
		},
	}

	s := server{}
	req := &synse.DeviceFilter{
		Rack:   "rack",
		Board:  "board",
		Device: "device",
	}
	mock := &test.MockReadStreamErr{}
	err := s.Read(req, mock)

	assert.Error(t, err)
}

// TestServer_Read4 tests the Read method of the gRPC plugin service when
// a bad device filter is specified.
func TestServer_Read4(t *testing.T) {
	s := server{}
	req := &synse.DeviceFilter{ // missing device to read
		Rack:  "rack",
		Board: "board",
	}
	mock := test.NewMockReadStream()
	err := s.Read(req, mock)

	assert.Error(t, err)
}

// TestServer_Write tests the Write method of the gRPC plugin service when
// the specified device isn't found.
func TestServer_Write(t *testing.T) {
	s := server{}
	req := &synse.WriteInfo{
		DeviceFilter: &synse.DeviceFilter{
			Rack:   "rack",
			Board:  "board",
			Device: "device",
		},
		Data: []*synse.WriteData{
			{Action: "test"},
		},
	}
	resp, err := s.Write(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestServer_Write2 tests the Write method of the gRPC plugin service
// when a bad device filter is specified.
func TestServer_Write2(t *testing.T) {
	s := server{}
	req := &synse.WriteInfo{
		DeviceFilter: &synse.DeviceFilter{ // missing device
			Rack:  "rack",
			Board: "board",
		},
		Data: []*synse.WriteData{
			{Action: "test"},
		},
	}
	resp, err := s.Write(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestServer_Write3 tests the Write method of the gRPC plugin service when
// there is only one WriteData specified.
func TestServer_Write3(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)
	defer func() {
		resetContext()
		Config.reset()
		DataManager = newDataManager()
	}()

	DataManager.writeChannel = make(chan *WriteContext, 20)
	Config.Plugin = &PluginConfig{
		Settings: &PluginSettings{
			Write: &WriteSettings{
				Enabled: true,
			},
		},
	}
	ctx.devices["rack-board-device"] = &Device{
		id:   "device",
		Kind: "foo",
		Location: &Location{
			Rack:  "rack",
			Board: "board",
		},
		Outputs: []*Output{
			{OutputType: OutputType{Name: "output1"}},
			{OutputType: OutputType{Name: "output2"}},
		},
		Handler: &DeviceHandler{
			Write: func(device *Device, data *WriteData) error {
				return nil
			},
		},
	}

	s := server{}
	req := &synse.WriteInfo{
		DeviceFilter: &synse.DeviceFilter{
			Rack:   "rack",
			Board:  "board",
			Device: "device",
		},
		Data: []*synse.WriteData{
			{Action: "test"},
		},
	}
	resp, err := s.Write(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(resp.Transactions))
}

// TestServer_Write4 tests the Write method of the gRPC plugin service when
// there are multiple write data specified.
func TestServer_Write4(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)
	defer func() {
		resetContext()
		Config.reset()
		DataManager = newDataManager()
	}()

	DataManager.writeChannel = make(chan *WriteContext, 20)
	Config.Plugin = &PluginConfig{
		Settings: &PluginSettings{
			Write: &WriteSettings{
				Enabled: true,
			},
		},
	}
	ctx.devices["rack-board-device"] = &Device{
		id:   "device",
		Kind: "foo",
		Location: &Location{
			Rack:  "rack",
			Board: "board",
		},
		Outputs: []*Output{
			{OutputType: OutputType{Name: "output1"}},
			{OutputType: OutputType{Name: "output2"}},
		},
		Handler: &DeviceHandler{
			Write: func(device *Device, data *WriteData) error {
				return nil
			},
		},
	}

	s := server{}
	req := &synse.WriteInfo{
		DeviceFilter: &synse.DeviceFilter{
			Rack:   "rack",
			Board:  "board",
			Device: "device",
		},
		Data: []*synse.WriteData{
			{Action: "foo"},
			{Action: "bar"},
		},
	}
	resp, err := s.Write(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(resp.Transactions))
}

// TestServer_Transaction tests the Transaction method of the gRPC plugin service.
func TestServer_Transaction(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)

	s := server{}
	req := &synse.TransactionFilter{}
	mock := test.NewMockTransactionStream()
	err := s.Transaction(req, mock)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(mock.Results))
}

// TestServer_Transaction2 tests the Transaction method of the gRPC plugin service
// when there are transactions in the cache and no filter.
func TestServer_Transaction2(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)

	t1 := newTransaction()
	t2 := newTransaction()

	s := server{}
	req := &synse.TransactionFilter{}
	mock := test.NewMockTransactionStream()
	err := s.Transaction(req, mock)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(mock.Results))
	assert.NotNil(t, mock.Results[t1.id])
	assert.NotNil(t, mock.Results[t2.id])
}

// TestServer_Transaction3 tests the Transaction method of the gRPC plugin service
// when there are transactions in the cache with a filter.
func TestServer_Transaction3(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)

	t1 := newTransaction()
	t2 := newTransaction()

	s := server{}
	req := &synse.TransactionFilter{Id: t1.id}
	mock := test.NewMockTransactionStream()
	err := s.Transaction(req, mock)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(mock.Results))
	assert.NotNil(t, mock.Results[t1.id])
	assert.Nil(t, mock.Results[t2.id])
}

// TestServer_Transaction4 tests the Transaction method of the gRPC plugin service
// when the filter does not match any transactions.
func TestServer_Transaction4(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)

	t1 := newTransaction()
	t2 := newTransaction()

	s := server{}
	req := &synse.TransactionFilter{Id: "abc"}
	mock := test.NewMockTransactionStream()
	err := s.Transaction(req, mock)

	assert.Error(t, err)
	assert.Equal(t, 0, len(mock.Results))
	assert.Nil(t, mock.Results[t1.id])
	assert.Nil(t, mock.Results[t2.id])
}

// TestServer_Transaction5 tests the Transaction method of the gRPC plugin service
// when sending the response results in error.
func TestServer_Transaction5(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)

	_ = newTransaction()
	_ = newTransaction()

	s := server{}
	req := &synse.TransactionFilter{}
	mock := &test.MockTransactionStreamErr{}
	err := s.Transaction(req, mock)

	assert.Error(t, err)
}
