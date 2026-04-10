// Package dlq provides a dead-letter queue for failed WAL events.
// Events that cannot be processed after all retry attempts are written
// to a configurable DLQ sink (file or in-memory) for later inspection.
package dlq

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/pgstream/pgstream/internal/wal"
)

// Entry represents a failed event stored in the dead-letter queue.
type Entry struct {
	Timestamp time.Time  `json:"timestamp"`
	Event     *wal.Event `json:"event"`
	Error     string     `json:"error"`
	Attempts  int        `json:"attempts"`
}

// DLQ holds failed events for later inspection or replay.
type DLQ struct {
	mu      sync.Mutex
	entries []Entry
	filePath string
}

// Config holds configuration for the dead-letter queue.
type Config struct {
	// FilePath is the optional path to persist DLQ entries as newline-delimited JSON.
	// If empty, entries are kept in memory only.
	FilePath string
}

// New creates a new DLQ with the given configuration.
func New(cfg Config) *DLQ {
	return &DLQ{
		filePath: cfg.FilePath,
	}
}

// Push adds a failed event to the dead-letter queue.
func (d *DLQ) Push(event *wal.Event, err error, attempts int) error {
	entry := Entry{
		Timestamp: time.Now().UTC(),
		Event:     event,
		Error:     err.Error(),
		Attempts:  attempts,
	}

	d.mu.Lock()
	d.entries = append(d.entries, entry)
	d.mu.Unlock()

	if d.filePath != "" {
		return d.persist(entry)
	}
	return nil
}

// Entries returns a snapshot of all dead-letter queue entries.
func (d *DLQ) Entries() []Entry {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]Entry, len(d.entries))
	copy(out, d.entries)
	return out
}

// Size returns the number of entries currently in the queue.
func (d *DLQ) Size() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.entries)
}

func (d *DLQ) persist(entry Entry) error {
	f, err := os.OpenFile(d.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("dlq: open file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	if err := enc.Encode(entry); err != nil {
		return fmt.Errorf("dlq: encode entry: %w", err)
	}
	return nil
}
