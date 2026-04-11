// Package eventlog provides a lightweight structured audit trail for pgstream.
//
// Each WAL event that passes through the pipeline can be recorded with its
// LSN, table, operation, and processing status (e.g. "sent", "filtered",
// "failed"). Entries are written as newline-delimited JSON to a file or to
// stdout when no path is configured.
//
// Usage:
//
//	log, err := eventlog.New(eventlog.Config{Path: "/var/log/pgstream/events.jsonl"})
//	if err != nil { ... }
//	defer log.Close()
//	log.Record(event, "sent", "")
package eventlog
