// Package acknowledge provides WAL LSN acknowledgement tracking for
// Postgres logical replication. It batches LSN confirmations and flushes
// them back to the replication connection at a configurable interval.
package acknowledge

import (
	"context"
	"sync"
	"time"
)

// Sender is the interface required to send standby status updates back to
// Postgres, confirming which LSN positions have been processed.
type Sender interface {
	SendStandbyStatusUpdate(ctx context.Context, lsn uint64) error
}

// Config holds configuration for the Acknowledger.
type Config struct {
	// FlushInterval controls how often pending LSNs are flushed. Defaults to 5s.
	FlushInterval time.Duration
}

func (c *Config) defaults() {
	if c.FlushInterval <= 0 {
		c.FlushInterval = 5 * time.Second
	}
}

// Acknowledger batches LSN positions and periodically confirms them to
// Postgres via the provided Sender.
type Acknowledger struct {
	mu      sync.Mutex
	cfg     Config
	sender  Sender
	pending uint64
	last    uint64
}

// New creates a new Acknowledger with the given Sender and Config.
func New(sender Sender, cfg Config) (*Acknowledger, error) {
	if sender == nil {
		return nil, ErrNilSender
	}
	cfg.defaults()
	return &Acknowledger{sender: sender, cfg: cfg}, nil
}

// Track records an LSN that has been successfully processed. Only advances
// the pending position if lsn is greater than the current pending value.
func (a *Acknowledger) Track(lsn uint64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if lsn > a.pending {
		a.pending = lsn
	}
}

// Flush sends the current pending LSN to Postgres if it has advanced since
// the last flush. It is safe to call concurrently.
func (a *Acknowledger) Flush(ctx context.Context) error {
	a.mu.Lock()
	pending := a.pending
	a.mu.Unlock()

	if pending <= a.last {
		return nil
	}
	if err := a.sender.SendStandbyStatusUpdate(ctx, pending); err != nil {
		return err
	}
	a.mu.Lock()
	a.last = pending
	a.mu.Unlock()
	return nil
}

// Run starts a background loop that flushes the pending LSN at the
// configured interval. It returns when ctx is cancelled.
func (a *Acknowledger) Run(ctx context.Context) error {
	ticker := time.NewTicker(a.cfg.FlushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			// Best-effort final flush.
			_ = a.Flush(context.Background()) //nolint:contextcheck
			return ctx.Err()
		case <-ticker.C:
			if err := a.Flush(ctx); err != nil {
				return err
			}
		}
	}
}
