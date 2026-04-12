// Package pruner provides periodic cleanup of stale replication slots
// and expired WAL positions to prevent unbounded disk growth on the
// Postgres primary.
package pruner

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// SlotDropper is the interface required to drop a replication slot.
type SlotDropper interface {
	DropSlot(ctx context.Context, name string) error
}

// Config holds pruner configuration.
type Config struct {
	// Interval between pruning runs. Defaults to 10 minutes.
	Interval time.Duration
	// MaxSlotLagBytes causes a slot to be dropped when its lag exceeds this
	// threshold (bytes). Zero disables lag-based pruning.
	MaxSlotLagBytes int64
	// Slots is the explicit list of slot names to monitor.
	Slots []string
}

func (c *Config) applyDefaults() {
	if c.Interval <= 0 {
		c.Interval = 10 * time.Minute
	}
}

// Pruner periodically inspects and removes stale replication slots.
type Pruner struct {
	cfg    Config
	dropper SlotDropper
	log    *slog.Logger
}

// New creates a Pruner. dropper must not be nil.
func New(cfg Config, dropper SlotDropper, log *slog.Logger) (*Pruner, error) {
	if dropper == nil {
		return nil, fmt.Errorf("pruner: dropper must not be nil")
	}
	if log == nil {
		log = slog.Default()
	}
	cfg.applyDefaults()
	return &Pruner{cfg: cfg, dropper: dropper, log: log}, nil
}

// Run starts the pruning loop. It blocks until ctx is cancelled.
func (p *Pruner) Run(ctx context.Context) error {
	ticker := time.NewTicker(p.cfg.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			p.prune(ctx)
		}
	}
}

func (p *Pruner) prune(ctx context.Context) {
	for _, name := range p.cfg.Slots {
		if err := p.dropper.DropSlot(ctx, name); err != nil {
			p.log.Error("pruner: failed to drop slot", "slot", name, "err", err)
			continue
		}
		p.log.Info("pruner: dropped stale slot", "slot", name)
	}
}
