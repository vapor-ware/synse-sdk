package sdk

import (
	"time"
)

// RWLoop is the read-write loop used by the PluginServer to perform device
// reads and writes in the background.
type RWLoop struct {
	handler        PluginHandler
	readingManager ReadingManager
	writingManager WritingManager
	devices        map[string]Device
}


// run the read-write loop.
//
// Device reads are pushed through the 'read' channel, where the ReadingManager
// should be listening on that channel and ready to receive. Once the
// ReadingManager gets the reading from the channel, it will be cached so the
// GRPC Read command can have access to it.
//
// Device writes happen asynchronously. When a GRPC Write command comes in, it
// is given a transaction id. That transaction is then queued up. At the start
// of every loop iteration, that queue is checked. If a transaction is present,
// it will be handled before the reads begin, otherwise the loop continues to
// read. To ensure that writes do not prevent reads from happening (e.g. if
// there are many many writes queued up), we will set a limit on the maximum
// number of write transactions that can be fulfilled per loop.
func (rwl *RWLoop) run() {

	go func() {
		loopDelay := Config.Settings.LoopDelay
		writesPerLoop := Config.Settings.Write.PerLoop

		for {
			// ~~ Write portion of the loop ~~
			// Get the next values off of the write queue
			for i := 0; i < writesPerLoop; i++ {
				select {
				case w := <- rwl.writingManager.channel:
					Logger.Debugf("Writing for %v (transaction: %v)", w.device, w.transaction.id)
					UpdateTransactionStatus(w.transaction.id, WRITING)

					data := writeDataFromGRPC(w.data)
					err := rwl.handler.Write(rwl.devices[w.device], data)
					if err != nil {
						UpdateTransactionState(w.transaction.id, ERROR)
						w.transaction.message = err.Error()
						Logger.Errorf("Failed to write to device %v: %v", w.device, err)
					}
					UpdateTransactionStatus(w.transaction.id, DONE)

				default:
					// if there is nothing in the write queue, there is nothing
					// to do here, so we just move on.
				}
			}


			// ~~ Read portion of the loop ~~
			for _, d := range rwl.devices {
				resp, err := rwl.handler.Read(d)
				if err != nil {
					// if there is an error reading, we will just log it and move on
					Logger.Errorf("Failed to read from device %v: %v", d.UID(), err)
				}
				rwl.readingManager.channel <- resp
			}

			// sleep for some configured amount of time. this isn't necessarily
			// needed, but it could be helpful to reduce resource usage and
			// keep things sane.
			if loopDelay != 0 {
				time.Sleep(time.Duration(loopDelay) * time.Millisecond)
			}
		}
	}()
	Logger.Info("[rwloop] Running")
}