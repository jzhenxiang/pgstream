// Package inspector provides a lightweight, thread-safe diagnostic view of
// the pgstream pipeline at runtime.
//
// An Inspector accumulates counters for WAL events (received, processed,
// failed) and tracks the highest LSN observed so far. It also maintains a
// boolean flag that reflects whether the downstream sink last responded
// successfully.
//
// Usage:
//
//	insp := inspector.New()
//	insp.RecordReceived(lsn.MustParse("0/1A2B3C"))
//	insp.RecordProcessed()
//	snap := insp.Snapshot()
//	fmt.Println(snap.LastLSN, snap.Processed)
//
// The Snapshot method returns a value copy, so callers can safely read it
// without holding any lock.
package inspector
