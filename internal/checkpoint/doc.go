// Package checkpoint provides a simple file-backed WAL LSN checkpoint store.
//
// It allows pgstream to resume WAL replication from the last successfully
// processed log sequence number (LSN) after a restart, avoiding duplicate
// or missed events.
//
// Usage:
//
//	cp, err := checkpoint.New("/var/lib/pgstream/checkpoint.json")
//	if err != nil { ... }
//
//	// After successfully processing a WAL event:
//	cp.Save(event.LSN)
//
//	// On startup, retrieve the last known position:
//	lastLSN := cp.Get()
package checkpoint
