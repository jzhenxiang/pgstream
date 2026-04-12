// Package aggregator groups WAL events by table and flushes them as batches
// once a configurable window size or time interval is reached.
package aggregator

import (
	"context"
	"sync"
	"time"

	"github.com/pgstream/pgstream/internal/wal"
)

// FlushFunc is called with the table name and accumulated events when a window
// is complete.
type FlushFunc func(ctx context.Context, table string, events []*wal.Event) error

// Aggregator collects events per table and flushes them in windows.
type Aggregator struct {
	mu           sync.Mutex
	buckets      map[string][]*wal.Event
	windowSize   int
	flushInterval time.Duration
	flush        FlushFunc
}

// New returns an Aggregator with the given configuration.
func New(cfg Config, flush FlushFunc) (*Aggregator, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	if flush == nil {
		return nil, errNilFlushFunc
	}
	return &Aggregator{
		buckets:       make(map[string][]*wal.Event),
		windowSize:    cfg.windowSize(),
		flushInterval: cfg.flushInterval(),
		flush:         flush,
	}, nil
}

// Add appends an event to the appropriate table bucket and flushes if the
// window size has been reached.
func (a *Aggregator) Add(ctx context.Context, event *wal.Event) error {
	if event == nil {
		return nil
	}
	a.mu.Lock()
	table := event.Table
	a.buckets[table] = append(a.buckets[table], event)
	ready := len(a.buckets[table]) >= a.windowSize
	var batch []*wal.Event
	if ready {
		batch = a.buckets[table]
		delete(a.buckets, table)
	}
	a.mu.Unlock()

	if ready {
		return a.flush(ctx, table, batch)
	}
	return nil
}

// Run starts a background ticker that flushes non-empty buckets periodically.
func (a *Aggregator) Run(ctx context.Context) error {
	ticker := time.NewTicker(a.flushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := a.flushAll(ctx); err != nil {
				return err
			}
		}
	}
}

func (a *Aggregator) flushAll(ctx context.Context) error {
	a.mu.Lock()
	snapshot := make(map[string][]*wal.Event, len(a.buckets))
	for t, evs := range a.buckets {
		snapshot[t] = evs
	}
	a.buckets = make(map[string][]*wal.Event)
	a.mu.Unlock()

	for table, evs := range snapshot {
		if err := a.flush(ctx, table, evs); err != nil {
			return err
		}
	}
	return nil
}
