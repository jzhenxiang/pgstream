// Package batcher implements size-and-time based batching for pgstream WAL events.
//
// A Batcher accumulates [wal.Event] values from a channel and forwards them
// to a [SendFunc] in bulk, either when the batch reaches MaxSize entries or
// when the FlushInterval ticker fires — whichever comes first.
//
// Usage:
//
//	b, err := batcher.New(batcher.Config{
//		MaxSize:       50,
//		FlushInterval: 2 * time.Second,
//	}, func(ctx context.Context, events []*wal.Event) error {
//		return sink.SendBatch(ctx, events)
//	})
//	if err != nil { ... }
//
//	err = b.Run(ctx, eventCh)
package batcher
