// Package slicer provides time-based and size-based window slicing for WAL event streams.
// It groups incoming events into discrete slices that can be flushed downstream.
package slicer

import (
	"errors"
	"sync"
	"time"

	"github.com/your-org/pgstream/internal/wal"
)

// DefaultMaxSize is the default maximum number of events per slice.
const DefaultMaxSize = 500

// DefaultInterval is the default flush interval.
const DefaultInterval = 5 * time.Second

// Config holds configuration for the Slicer.
type Config struct {
	MaxSize  int
	Interval time.Duration
}

func (c *Config) setDefaults() {
	if c.MaxSize <= 0 {
		c.MaxSize = DefaultMaxSize
	}
	if c.Interval <= 0 {
		c.Interval = DefaultInterval
	}
}

// Slicer accumulates events and flushes them as slices when either
// the size limit or the time interval is reached.
type Slicer struct {
	cfg    Config
	buf    []*wal.Event
	mu     sync.Mutex
	flush  func([]*wal.Event) error
}

// New creates a new Slicer with the given config and flush callback.
// flush is called each time a slice is ready.
func New(cfg Config, flush func([]*wal.Event) error) (*Slicer, error) {
	if flush == nil {
		return nil, errors.New("slicer: flush function must not be nil")
	}
	cfg.setDefaults()
	return &Slicer{
		cfg:   cfg,
		buf:   make([]*wal.Event, 0, cfg.MaxSize),
		flush: flush,
	}, nil
}

// Add appends an event to the current slice. If the slice reaches MaxSize,
// it is flushed immediately.
func (s *Slicer) Add(event *wal.Event) error {
	if event == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buf = append(s.buf, event)
	if len(s.buf) >= s.cfg.MaxSize {
		return s.flushLocked()
	}
	return nil
}

// Flush forces an immediate flush of any buffered events.
func (s *Slicer) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.flushLocked()
}

func (s *Slicer) flushLocked() error {
	if len(s.buf) == 0 {
		return nil
	}
	slice := make([]*wal.Event, len(s.buf))
	copy(slice, s.buf)
	s.buf = s.buf[:0]
	return s.flush(slice)
}
