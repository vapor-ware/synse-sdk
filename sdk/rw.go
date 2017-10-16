package sdk

import (
	"fmt"
	"time"
)

//
type RWLoop struct {
	handler PluginHandler

	readingManager ReadingManager
	writingManager WritingManager

	devices map[string]Device
}


// This is the Read-Write loop. It is where the plugin will perform its
// background device reads and writes.
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
func (rwl *RWLoop) Run() {

	go func() {
		loopDelay := Config.LoopDelay
		writesPerLoop := Config.WritesPerLoop

		for {
			// ~~ Write portion of the loop ~~
			// Get the next values off of the write queue
			for i := 0; i < writesPerLoop; i++ {
				select {
				case w := <- rwl.writingManager.channel:

					fmt.Printf("Write for: %v\n", w)

					UpdateTransactionStatus(w.transaction.id, WRITING)

					err := rwl.handler.Write(rwl.devices[w.device], w.data)
					if err != nil {
						UpdateTransactionState(w.transaction.id, ERROR)
						fmt.Errorf("Error when writing device: %v\n", err)
					}
					UpdateTransactionStatus(w.transaction.id, DONE)
					fmt.Printf("-- tid: %v\n", w.transaction.id)

				default:
					// if there is nothing in the write queue, there is nothing
					// to do here, so we just move on.
				}
			}


			// ~~ Read portion of the loop ~~
			for _, d := range rwl.devices {
				resp, err := rwl.handler.Read(d)
				if err != nil {
					// if there is an error reading, we will just log it and move on for now
					// TODO - determine what the desired behavior is. this might be correct,
					// but need to think through the logging/alerting/etc scenarios.
					fmt.Errorf("Error when reading device: %v\n", err)
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
	logger.Info("[rwloop] running")
}