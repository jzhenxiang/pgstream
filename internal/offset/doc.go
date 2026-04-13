// Package offset provides a lightweight WAL offset tracker for pgstream.
//
// It records the last successfully processed Log Sequence Number (LSN) both
// in memory and on disk, allowing the replication pipeline to resume from
// the correct position after a process restart without reprocessing already
// delivered events.
//
// The tracker is safe for concurrent use. Commits are written atomically to
// disk using a write-and-rename strategy to avoid partial writes on crash.
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
//
// If the offset file does not exist when New is called, the tracker starts
// from zero and will create the file on the first successful Commit.
package offset
