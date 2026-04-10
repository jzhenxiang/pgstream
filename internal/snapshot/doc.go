// Package snapshot implements initial table snapshotting for pgstream.
//
// Before streaming incremental WAL changes it is often useful to capture
// the current state of one or more tables. The Snapshot type connects to
// Postgres, reads every row from the configured tables in batches and
// emits each row as a synthetic wal.Event so it travels through the
// normal filter → transform → sink pipeline unchanged.
//
// Usage:
//
//	snap, err := snapshot.New(snapshot.Config{
//	    DSN:    "postgres://...",
//	    Tables: []string{"public.orders", "public.users"},
//	})
//	if err != nil { ... }
//	if err := snap.Run(ctx, pipeline.Emit); err != nil { ... }
package snapshot
