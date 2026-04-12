// Package slicer groups WAL events into discrete time-based or size-based
// slices before forwarding them to a downstream flush function.
//
// A Slicer accumulates events in an internal buffer and flushes them when
// either the configured maximum slice size is reached or the flush interval
// elapses. This provides natural batching without coupling the upstream
// producer to downstream throughput.
//
// Example usage:
//
//	s, _ := slicer.New(slicer.Config{MaxSize: 200, Interval: 2 * time.Second}, myFlush)
//	s.Add(event)
package slicer
