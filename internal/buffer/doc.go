// Package buffer implements a size- and time-based batching buffer for WAL
// events. It accumulates incoming events and flushes them in batches either
// when a configurable size threshold is reached or when a periodic flush
// interval elapses — whichever comes first.
//
// Usage:
//
//	buf := buffer.New(buffer.Config{
//		MaxSize:       50,
//		FlushInterval: 2 * time.Second,
//	}, func(events []*wal.Event) error {
//		return sink.Send(events)
//	})
//
//	go buf.Run(ctx)
//
//	for _, e := range events {
//		buf.Add(e)
//	}
package buffer
