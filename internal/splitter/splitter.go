// Package splitter provides a WAL event splitter that divides a batch of
// events into smaller chunks for downstream processing.
package splitter

import (
	"errors"

	"github.com/pgstream/pgstream/internal/wal"
)

const defaultChunkSize = 100

// Splitter divides a slice of WAL events into fixed-size chunks.
type Splitter struct {
	chunkSize int
}

// Config holds configuration for the Splitter.
type Config struct {
	// ChunkSize is the maximum number of events per chunk.
	// Defaults to 100 if zero.
	ChunkSize int
}

// New creates a new Splitter with the provided configuration.
// Returns an error if ChunkSize is negative.
func New(cfg Config) (*Splitter, error) {
	if cfg.ChunkSize < 0 {
		return nil, errors.New("splitter: chunk size must be non-negative")
	}
	size := cfg.ChunkSize
	if size == 0 {
		size = defaultChunkSize
	}
	return &Splitter{chunkSize: size}, nil
}

// Split partitions events into chunks of at most ChunkSize elements.
// Returns an empty slice if events is nil or empty.
func (s *Splitter) Split(events []*wal.Event) [][]*wal.Event {
	if len(events) == 0 {
		return [][]*wal.Event{}
	}

	var chunks [][]*wal.Event
	for i := 0; i < len(events); i += s.chunkSize {
		end := i + s.chunkSize
		if end > len(events) {
			end = len(events)
		}
		chunk := make([]*wal.Event, end-i)
		copy(chunk, events[i:end])
		chunks = append(chunks, chunk)
	}
	return chunks
}

// ChunkSize returns the configured chunk size.
func (s *Splitter) ChunkSize() int {
	return s.chunkSize
}
