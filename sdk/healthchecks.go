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

//// readBufferHealthCheck is a plugin health check that looks at the data manager's
//// read buffer and determines whether the buffer is becoming full or not. If the
//// read buffer is >95% full, it will cause the health check to fail.
//func readBufferHealthCheck() error {
//	// current number of elements queued in the channel
//	channelSize := len(DataManager.readChannel)
//
//	// total capacity of the channel
//	channelCap := cap(DataManager.readChannel)
//
//	pctUsage := (float64(channelSize) / float64(channelCap)) * 100
//	if pctUsage > 95 {
//		return fmt.Errorf("data manager read buffer usage >95%%, consider increasing buffer size in plugin config")
//	}
//	return nil
//}
//
//// writeBufferHealthCheck is a plugin health check that looks at the data manager's
//// write buffer and determines whether the buffer is becoming full or not. If the
//// write buffer is >95% full, it will cause the health check to fail.
//func writeBufferHealthCheck() error {
//	// current number of elements queued in the channel
//	channelSize := len(DataManager.writeChannel)
//
//	// total capacity of the channel
//	channelCap := cap(DataManager.writeChannel)
//
//	pctUsage := (float64(channelSize) / float64(channelCap)) * 100
//	if pctUsage > 95 {
//		return fmt.Errorf("data manager write buffer usage >95%%, consider increasing buffer size in plugin config")
//	}
//	return nil
//}
