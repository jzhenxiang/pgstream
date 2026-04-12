// Package cursor provides a thread-safe LSN cursor for tracking the current
// read position within a PostgreSQL WAL stream.
//
// A Cursor wraps a single Log Sequence Number (LSN) and exposes Advance and
// Reset operations that are safe for concurrent use. It also maintains a
// high-water mark so callers can distinguish the furthest position ever seen
// from the most recently committed one.
//
// Typical usage:
//
//	c := cursor.New(0)
//	c.Advance(lsn)
//	fmt.Println(c.Current())
package cursor
