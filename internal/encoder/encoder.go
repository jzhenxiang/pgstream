// Package encoder provides JSON encoding utilities for WAL events
// before they are dispatched to sinks.
package encoder

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pgstream/pgstream/internal/wal"
)

// Envelope wraps a WAL event with metadata for downstream consumers.
type Envelope struct {
	Version   int            `json:"version"`
	Timestamp time.Time      `json:"timestamp"`
	Event     *wal.Event     `json:"event"`
	Meta      map[string]any `json:"meta,omitempty"`
}

// Encoder encodes WAL events into JSON bytes.
type Encoder struct {
	version int
}

// New returns a new Encoder.
func New() *Encoder {
	return &Encoder{version: 1}
}

// Encode wraps the event in an Envelope and marshals it to JSON.
func (e *Encoder) Encode(event *wal.Event) ([]byte, error) {
	if event == nil {
		return nil, fmt.Errorf("encoder: nil event")
	}

	env := Envelope{
		Version:   e.version,
		Timestamp: time.Now().UTC(),
		Event:     event,
	}

	b, err := json.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("encoder: marshal: %w", err)
	}

	return b, nil
}

// EncodeWithMeta is like Encode but attaches arbitrary metadata to the envelope.
func (e *Encoder) EncodeWithMeta(event *wal.Event, meta map[string]any) ([]byte, error) {
	if event == nil {
		return nil, fmt.Errorf("encoder: nil event")
	}

	env := Envelope{
		Version:   e.version,
		Timestamp: time.Now().UTC(),
		Event:     event,
		Meta:      meta,
	}

	b, err := json.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("encoder: marshal with meta: %w", err)
	}

	return b, nil
}
