package sdk

import (
	"fmt"
)

// readBufferHealthCheck is a plugin health check that looks at the data manager's
// read buffer and determines whether the buffer is becoming full or not. If the
// read buffer is >95% full, it will cause the health check to fail.
func readBufferHealthCheck() error {
	// current number of elements queued in the channel
	channelSize := len(DataManager.readChannel)

	// total capacity of the channel
	channelCap := cap(DataManager.readChannel)

	pctUsage := float64(channelSize) / float64(channelCap)
	if pctUsage > 95 {
		return fmt.Errorf("data manager read buffer usage >95%%, consider increasing buffer size in plugin config")
	}
	return nil
}

// writeBufferHealthCheck is a plugin health check that looks at the data manager's
// write buffer and determines whether the buffer is becoming full or not. If the
// write buffer is >95% full, it will cause the health check to fail.
func writeBufferHealthCheck() error {
	// current number of elements queued in the channel
	channelSize := len(DataManager.writeChannel)

	// total capacity of the channel
	channelCap := cap(DataManager.writeChannel)

	pctUsage := float64(channelSize) / float64(channelCap)
	if pctUsage > 95 {
		return fmt.Errorf("data manager write buffer usage >95%%, consider increasing buffer size in plugin config")
	}
	return nil
}
