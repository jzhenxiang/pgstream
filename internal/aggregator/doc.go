// Package aggregator provides a table-scoped event aggregator for pgstream.
//
// Events arriving from the WAL are grouped by their table name. A flush is
// triggered automatically when either:
//
//   - the number of buffered events for a table reaches WindowSize, or
//   - the FlushInterval ticker fires (whichever comes first).
//
// Typical usage:
//
//	agg, err := aggregator.New(aggregator.Config{WindowSize: 50}, myFlush)
//	if err != nil { ... }
//	go agg.Run(ctx)
//	for _, ev := range events {
//	    agg.Add(ctx, ev)
//	}
package aggregator
