// Package backpressure implements a semaphore-based backpressure controller
// for the pgstream pipeline.
//
// When a downstream sink (Kafka or webhook) cannot keep up with the rate of
// incoming WAL events the Controller will block the reader from acquiring new
// slots. Once the number of pending events exceeds MaxPending the caller
// receives ErrBackpressure, signalling that the pipeline should pause or
// slow down ingestion.
//
// Typical usage:
//
//	ctrl := backpressure.New(backpressure.DefaultConfig())
//	if err := ctrl.Acquire(ctx); err != nil {
//	    // handle slow sink
//	}
//	defer ctrl.Release()
package backpressure
