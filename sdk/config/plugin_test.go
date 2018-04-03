package config

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// TestSettings_IsParallel tests checking whether a Settings instance
// is configured for parallel mode.
func TestSettings_IsParallel(t *testing.T) {
	s := Settings{}

	s.Mode = "serial"
	assert.False(t, s.IsParallel())

	s.Mode = "parallel"
	assert.True(t, s.IsParallel())
}

// TestSettings_IsSerial tests checking whether a Settings instance
// is configured for serial mode.
func TestSettings_IsSerial(t *testing.T) {
	s := Settings{}

	s.Mode = "serial"
	assert.True(t, s.IsSerial())

	s.Mode = "parallel"
	assert.False(t, s.IsSerial())
}

// TestReadSettings_GetInterval tests getting the read interval setting
// successfully.
func TestReadSettings_GetInterval(t *testing.T) {
	s := ReadSettings{}
	s.Interval = "10s"

	interval, err := s.GetInterval()
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(10)*time.Second, interval)
}

// TestReadSettings_GetIntervalErr tests getting the read interval setting
// unsuccessfully.
func TestReadSettings_GetIntervalErr(t *testing.T) {
	s := ReadSettings{}
	s.Interval = "abc"

	interval, err := s.GetInterval()
	assert.Equal(t, time.Duration(0), interval)
	assert.Error(t, err)
}

// TestWriteSettings_GetInterval tests getting the write interval setting
// successfully.
func TestWriteSettings_GetInterval(t *testing.T) {
	s := WriteSettings{}
	s.Interval = "10s"

	interval, err := s.GetInterval()
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(10)*time.Second, interval)
}

// TestWriteSettings_GetIntervalErr tests getting the write interval setting
// unsuccessfully.
func TestWriteSettings_GetIntervalErr(t *testing.T) {
	s := WriteSettings{}
	s.Interval = "abc"

	interval, err := s.GetInterval()
	assert.Equal(t, time.Duration(0), interval)
	assert.Error(t, err)
}

// TestNewPluginConfigErr tests creating a new plugin config that should result
// in an error because no plugin configuration file can be found.
func TestNewPluginConfigErr(t *testing.T) {
	config, err := NewPluginConfig()
	assert.Error(t, err)
	assert.Nil(t, config)
}

// TestPluginConfig_Validate tests validating the PluginConfig successfully.
func TestPluginConfig_Validate(t *testing.T) {
	var cases = []PluginConfig{
		{
			Name:    "test",
			Version: "1",
			Network: NetworkSettings{
				Type:    "test",
				Address: "test",
			},
			Settings: Settings{
				Transaction: TransactionSettings{
					TTL: "300s",
				},
			},
		},
		{
			Name:    "1",
			Version: "2",
			Network: NetworkSettings{
				Type:    "3",
				Address: "4",
			},
			Settings: Settings{
				Transaction: TransactionSettings{
					TTL: "20m",
				},
			},
		},
		{
			Name:    "test",
			Version: "1",
			Network: NetworkSettings{
				Type:    "test",
				Address: "test",
			},
			Settings: Settings{
				Transaction: TransactionSettings{
					TTL: "0s",
				},
			},
		},
	}

	for _, testCase := range cases {
		err := testCase.Validate()
		assert.NoError(t, err)
	}
}

// TestPluginConfig_ValidateErr tests validating the PluginConfig unsuccessfully.
func TestPluginConfig_ValidateErr(t *testing.T) {
	var cases = []PluginConfig{
		{
			Version: "1",
			Network: NetworkSettings{
				Type:    "test",
				Address: "test",
			},
			Settings: Settings{
				Transaction: TransactionSettings{
					TTL: "2s",
				},
			},
		},
		{
			Name: "test",
			Network: NetworkSettings{
				Type:    "test",
				Address: "test",
			},
			Settings: Settings{
				Transaction: TransactionSettings{
					TTL: "2s",
				},
			},
		},
		{
			Name:    "test",
			Version: "1",
			Settings: Settings{
				Transaction: TransactionSettings{
					TTL: "2s",
				},
			},
		},
		{
			Name:    "test",
			Version: "1",
			Network: NetworkSettings{
				Address: "test",
			},
			Settings: Settings{
				Transaction: TransactionSettings{
					TTL: "2s",
				},
			},
		},
		{
			Name:    "test",
			Version: "1",
			Network: NetworkSettings{
				Type: "test",
			},
			Settings: Settings{
				Transaction: TransactionSettings{
					TTL: "2s",
				},
			},
		},
		{
			Name:    "1",
			Version: "2",
			Network: NetworkSettings{
				Type:    "3",
				Address: "4",
			},
			Settings: Settings{
				Transaction: TransactionSettings{
					TTL: "not-a-duration",
				},
			},
		},
	}

	for _, testCase := range cases {
		err := testCase.Validate()
		assert.Error(t, err)
	}
}

// TestSetupLookupInfo tests setting up the lookup info for a Viper instance
// for the plugin configuration. setLookupInfo does not return anything, and
// the values it sets on the Viper instance are not exported, so we just make
// sure nothing goes wrong here.
func TestSetLookupInfo(t *testing.T) {
	v := viper.New()
	setLookupInfo(v)
}
