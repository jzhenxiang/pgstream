// Package sequencer assigns monotonically increasing sequence numbers to WAL
// events so downstream consumers can detect gaps or reordering.
package sequencer

import (
	"errors"
	"sync/atomic"

	"pgstream/internal/wal"
)

// Sequencer stamps each event with a uint64 sequence number that starts at 1
// and increments by 1 for every call to Next.
type Sequencer struct {
	counter atomic.Uint64
	field   string
}

// Config holds options for the Sequencer.
type Config struct {
	// Field is the metadata key under which the sequence number is stored.
	// Defaults to "_seq" when empty.
	Field string
}

var errNilEvent = errors.New("sequencer: event must not be nil")

// New creates a Sequencer using the provided Config.
func New(cfg Config) *Sequencer {
	field := cfg.Field
	if field == "" {
		field = "_seq"
	}
	return &Sequencer{field: field}
}

// Next stamps event.Metadata with the next sequence number and returns the
// event. It returns errNilEvent when event is nil.
func (s *Sequencer) Next(event *wal.Event) (*wal.Event, error) {
	if event == nil {
		return nil, errNilEvent
	}

	seq := s.counter.Add(1)

	if event.Metadata == nil {
		event.Metadata = make(map[string]any)
	}
	event.Metadata[s.field] = seq

	return event, nil
}

// Current returns the last sequence number that was issued. Zero means no
// event has been stamped yet.
func (s *Sequencer) Current() uint64 {
	return s.counter.Load()
}

// Reset sets the internal counter back to zero. Useful in tests.
func (s *Sequencer) Reset() {
	s.counter.Store(0)
}
