// Package heartbeat provides periodic heartbeat events to keep WAL replication
// slots alive during periods of low activity.
package heartbeat

import (
	"context"
	"log"
	"time"
)

// DefaultInterval is the default heartbeat interval.
const DefaultInterval = 10 * time.Second

// Sender is anything that can send a heartbeat signal to Postgres.
type Sender interface {
	SendStandbyStatus(ctx context.Context) error
}

// Config holds heartbeat configuration.
type Config struct {
	// Interval between heartbeat signals. Defaults to DefaultInterval.
	Interval time.Duration
	// Logger is optional; uses the standard logger when nil.
	Logger *log.Logger
}

// Heartbeat sends periodic standby status updates to keep the replication
// slot alive.
type Heartbeat struct {
	cfg    Config
	sender Sender
	logger *log.Logger
}

// New creates a new Heartbeat. interval <= 0 falls back to DefaultInterval.
func New(sender Sender, cfg Config) (*Heartbeat, error) {
	if sender == nil {
		return nil, errNilSender
	}
	if cfg.Interval <= 0 {
		cfg.Interval = DefaultInterval
	}
	logger := cfg.Logger
	if logger == nil {
		logger = log.Default()
	}
	return &Heartbeat{cfg: cfg, sender: sender, logger: logger}, nil
}

// Run starts sending heartbeats at the configured interval until ctx is done.
func (h *Heartbeat) Run(ctx context.Context) error {
	ticker := time.NewTicker(h.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := h.sender.SendStandbyStatus(ctx); err != nil {
				h.logger.Printf("heartbeat: send standby status: %v", err)
			}
		}
	}
}
