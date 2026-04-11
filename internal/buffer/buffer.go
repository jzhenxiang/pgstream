// Package buffer provides a batching buffer that accumulates WAL events
// and flushes them based on size or time thresholds.
package buffer

import (
	"context"
	"sync"
	"time"

	"github.com/pgstream/pgstream/internal/wal"
)

// DefaultMaxSize is the default number of events before a flush is triggered.
const DefaultMaxSize = 100

// DefaultFlushInterval is the default time between forced flushes.
const DefaultFlushInterval = 5 * time.Second

// Config holds configuration for the Buffer.
type Config struct {
	MaxSize       int
	FlushInterval time.Duration
}

// Buffer accumulates WAL events and flushes them in batches.
type Buffer struct {
	mu       sync.Mutex
	events   []*wal.Event
	cfg      Config
	flushFn  func([]*wal.Event) error
}

// New creates a new Buffer with the given config and flush callback.
func New(cfg Config, flushFn func([]*wal.Event) error) *Buffer {
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = DefaultMaxSize
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = DefaultFlushInterval
	}
	return &Buffer{
		events:  make([]*wal.Event, 0, cfg.MaxSize),
		cfg:     cfg,
		flushFn: flushFn,
	}
}

// Add appends an event to the buffer, flushing if the size threshold is reached.
func (b *Buffer) Add(event *wal.Event) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.events = append(b.events, event)
	if len(b.events) >= b.cfg.MaxSize {
		return b.flush()
	}
	return nil
}

// Run starts the periodic flush loop, blocking until ctx is cancelled.
func (b *Buffer) Run(ctx context.Context) error {
	ticker := time.NewTicker(b.cfg.FlushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			b.mu.Lock()
			defer b.mu.Unlock()
			if len(b.events) > 0 {
				_ = b.flush()
			}
			return ctx.Err()
		case <-ticker.C:
			b.mu.Lock()
			if len(b.events) > 0 {
				if err := b.flush(); err != nil {
					b.mu.Unlock()
					return err
				}
			}
			b.mu.Unlock()
		}
	}
}

// Len returns the current number of buffered events.
func (b *Buffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.events)
}

// flush sends buffered events to the flushFn and resets the buffer.
// Caller must hold b.mu.
func (b *Buffer) flush() error {
	batch := make([]*wal.Event, len(b.events))
	copy(batch, b.events)
	b.events = b.events[:0]
	return b.flushFn(batch)
}
