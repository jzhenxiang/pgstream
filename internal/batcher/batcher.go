// Package batcher provides time-and-size based batching of WAL events
// before forwarding them to a sink. It accumulates events until either
// the batch reaches MaxSize or the FlushInterval elapses.
package batcher

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/pgstream/pgstream/internal/wal"
)

// ErrNoSink is returned when no send function is provided.
var ErrNoSink = errors.New("batcher: send function must not be nil")

// SendFunc is a function that processes a batch of events.
type SendFunc func(ctx context.Context, events []*wal.Event) error

// Config holds tunable parameters for the Batcher.
type Config struct {
	MaxSize       int
	FlushInterval time.Duration
}

func (c *Config) applyDefaults() {
	if c.MaxSize <= 0 {
		c.MaxSize = 100
	}
	if c.FlushInterval <= 0 {
		c.FlushInterval = 5 * time.Second
	}
}

// Batcher accumulates events and flushes them in batches.
type Batcher struct {
	cfg    Config
	send   SendFunc
	mu     sync.Mutex
	batch  []*wal.Event
}

// New creates a new Batcher with the given config and send function.
func New(cfg Config, send SendFunc) (*Batcher, error) {
	if send == nil {
		return nil, ErrNoSink
	}
	cfg.applyDefaults()
	return &Batcher{
		cfg:   cfg,
		send:  send,
		batch: make([]*wal.Event, 0, cfg.MaxSize),
	}, nil
}

// Run starts the batcher loop. It flushes on tick or when MaxSize is reached.
// It blocks until ctx is cancelled.
func (b *Batcher) Run(ctx context.Context, events <-chan *wal.Event) error {
	ticker := time.NewTicker(b.cfg.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			b.flush(ctx)
			return ctx.Err()
		case ev, ok := <-events:
			if !ok {
				b.flush(ctx)
				return nil
			}
			b.mu.Lock()
			b.batch = append(b.batch, ev)
			ready := len(b.batch) >= b.cfg.MaxSize
			b.mu.Unlock()
			if ready {
				if err := b.flush(ctx); err != nil {
					return err
				}
			}
		case <-ticker.C:
			if err := b.flush(ctx); err != nil {
				return err
			}
		}
	}
}

func (b *Batcher) flush(ctx context.Context) error {
	b.mu.Lock()
	if len(b.batch) == 0 {
		b.mu.Unlock()
		return nil
	}
	toSend := b.batch
	b.batch = make([]*wal.Event, 0, b.cfg.MaxSize)
	b.mu.Unlock()
	return b.send(ctx, toSend)
}
