// Package acknowledge provides periodic WAL LSN acknowledgement for Postgres
// logical replication slots.
//
// # Overview
//
// During logical replication Postgres expects the client to periodically
// confirm which log sequence numbers (LSNs) have been safely consumed.
// Failing to do so causes the replication slot to retain WAL segments
// indefinitely, eventually filling the server's disk.
//
// The Acknowledger batches processed LSNs in memory and flushes the
// highest confirmed position back to Postgres at a configurable interval
// via the Sender interface.
//
// # Usage
//
//	ack, err := acknowledge.New(conn, acknowledge.Config{FlushInterval: 2 * time.Second})
//	go ack.Run(ctx)
//	// After processing each WAL event:
//	ack.Track(event.LSN)
package acknowledge
