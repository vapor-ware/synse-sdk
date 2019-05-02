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

package config

import "testing"

//
// Unless there are updates done to the logger such that we can
// capture the logged info, we will only test that the logging functions
// run and don't crash anything.
//
// Each config struct is tested with a nil copy and initialized but
// empty copy.
//

func TestPlugin_Log_nil(t *testing.T) {
	var c *Plugin
	c.Log()
}

func TestPlugin_Log(t *testing.T) {
	c := Plugin{}
	c.Log()
}

func TestIDSettings_Log_nil(t *testing.T) {
	var c *IDSettings
	c.Log()
}

func TestIDSettings_Log(t *testing.T) {
	c := IDSettings{}
	c.Log()
}

func TestMetricsSettings_Log_nil(t *testing.T) {
	var c *MetricsSettings
	c.Log()
}

func TestMetricsSettings_Log(t *testing.T) {
	c := MetricsSettings{}
	c.Log()
}

func TestPluginSettings_Log_nil(t *testing.T) {
	var c *PluginSettings
	c.Log()
}

func TestPluginSettings_Log(t *testing.T) {
	c := PluginSettings{}
	c.Log()
}

func TestListenSettings_Log_nil(t *testing.T) {
	var c *ListenSettings
	c.Log()
}

func TestListenSettings_Log(t *testing.T) {
	c := ListenSettings{}
	c.Log()
}

func TestReadSettings_Log_nil(t *testing.T) {
	var c *ReadSettings
	c.Log()
}

func TestReadSettings_Log(t *testing.T) {
	c := ReadSettings{}
	c.Log()
}

func TestWriteSettings_Log_nil(t *testing.T) {
	var c *WriteSettings
	c.Log()
}

func TestWriteSettings_Log(t *testing.T) {
	c := WriteSettings{}
	c.Log()
}

func TestTransactionSettings_Log_nil(t *testing.T) {
	var c *TransactionSettings
	c.Log()
}

func TestTransactionSettings_Log(t *testing.T) {
	c := TransactionSettings{}
	c.Log()
}

func TestLimiterSettings_Log_nil(t *testing.T) {
	var c *LimiterSettings
	c.Log()
}

func TestLimiterSettings_Log(t *testing.T) {
	c := LimiterSettings{}
	c.Log()
}

func TestCacheSettings_Log_nil(t *testing.T) {
	var c *CacheSettings
	c.Log()
}

func TestCacheSettings_Log(t *testing.T) {
	c := CacheSettings{}
	c.Log()
}

func TestNetworkSettings_Log_nil(t *testing.T) {
	var c *NetworkSettings
	c.Log()
}

func TestNetworkSettings_Log(t *testing.T) {
	c := NetworkSettings{}
	c.Log()
}

func TestTLSNetworkSettings_Log_nil(t *testing.T) {
	var c *TLSNetworkSettings
	c.Log()
}

func TestTLSNetworkSettings_Log(t *testing.T) {
	c := TLSNetworkSettings{}
	c.Log()
}

func TestDynamicRegistrationSettings_Log_nil(t *testing.T) {
	var c *DynamicRegistrationSettings
	c.Log()
}

func TestDynamicRegistrationSettings_Log(t *testing.T) {
	c := DynamicRegistrationSettings{}
	c.Log()
}

func TestHealthSettings_Log_nil(t *testing.T) {
	var c *HealthSettings
	c.Log()
}
func TestHealthSettings_Log(t *testing.T) {
	c := HealthSettings{}
	c.Log()
}

func TestHealthCheckSettings_Log_nil(t *testing.T) {
	var c *HealthCheckSettings
	c.Log()
}

func TestHealthCheckSettings_Log(t *testing.T) {
	c := HealthCheckSettings{}
	c.Log()
}
