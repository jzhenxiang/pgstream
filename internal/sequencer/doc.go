// Package sequencer provides a thread-safe monotonic sequence number stamper
// for WAL events.
//
// Usage:
//
//	seq := sequencer.New(sequencer.Config{})
//	stampedEvent, err := seq.Next(event)
//
// The sequence number is stored in event.Metadata under the configured field
// name (default "_seq"). Downstream consumers can use this value to detect
// gaps caused by filtering, deduplication, or network issues.
package sequencer
