// Package dlq implements a dead-letter queue (DLQ) for pgstream.
//
// When a WAL event fails processing after all configured retry attempts,
// it is handed off to the DLQ so that no data is silently dropped.
// Entries can be stored in memory for testing or flushed to a newline-
// delimited JSON file for durable persistence and offline replay.
//
// Usage:
//
//	q := dlq.New(dlq.Config{FilePath: "/var/log/pgstream/dlq.jsonl"})
//	if err := q.Push(event, processingErr, attempts); err != nil {
//		log.Printf("dlq write error: %v", err)
//	}
package dlq
