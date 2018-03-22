package config

import (
	"testing"
)

var pluginConfigValidateTestList = []PluginConfig{
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
}

func TestPluginConfig_Validate(t *testing.T) {
	for _, ti := range pluginConfigValidateTestList {
		err := ti.Validate()
		if err != nil {
			t.Errorf("PluginConfig.Validate() expected no error but got %v", err)
		}
	}
}

var pluginConfigValidateErrorsTestList = []PluginConfig{
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

func TestPluginConfig_Validate2(t *testing.T) {
	for _, ti := range pluginConfigValidateErrorsTestList {
		err := ti.Validate()
		if err == nil {
			t.Error("PluginConfig.Validate() expected error but got nil.")
		}
	}
}
