// Package offset provides a lightweight WAL offset tracker for pgstream.
//
// It records the last successfully processed Log Sequence Number (LSN) both
// in memory and on disk, allowing the replication pipeline to resume from
// the correct position after a process restart without reprocessing already
// delivered events.
//
// Usage:
//
//	tracker, err := offset.New("/var/lib/pgstream/offset.json")
//	if err != nil { ... }
//
//	// after a batch is delivered
//	if err := tracker.Commit(lsn); err != nil { ... }
//
//	// on startup, retrieve the resume point
//	fmt.Println(tracker.Current())
package offset
