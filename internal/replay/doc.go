// Package replay provides WAL event replay functionality for pgstream.
//
// It reads WAL messages from a Reader, forwards them to a Sink, and commits
// the LSN position to an Offset store after each successful send. On restart,
// the Replayer resumes from the last committed position, ensuring at-least-once
// delivery semantics.
//
// Typical usage:
//
//	off, _ := offset.New(offset.Config{FilePath: "/var/lib/pgstream/offset"})
//	r, _ := replay.New(walReader, kafkaSink, off, logger)
//	r.Run(ctx)
package replay
