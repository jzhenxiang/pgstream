// Package backpressure provides a simple token-bucket style backpressure
// mechanism to throttle WAL event processing when downstream sinks are slow.
package backpressure

import (
	"context"
	"sync"
	"time"
)

// Config holds configuration for the backpressure controller.
type Config struct {
	// MaxPending is the maximum number of unacknowledged events allowed.
	MaxPending int
	// AcquireTimeout is how long Acquire will wait before returning an error.
	AcquireTimeout time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		MaxPending:     256,
		AcquireTimeout: 5 * time.Second,
	}
}

// Controller limits concurrent in-flight events using a buffered semaphore.
type Controller struct {
	sem     chan struct{}
	mu      sync.Mutex
	pending int
	cfg     Config
}

// New creates a new Controller with the given Config.
func New(cfg Config) *Controller {
	if cfg.MaxPending <= 0 {
		cfg.MaxPending = DefaultConfig().MaxPending
	}
	if cfg.AcquireTimeout <= 0 {
		cfg.AcquireTimeout = DefaultConfig().AcquireTimeout
	}
	return &Controller{
		sem: make(chan struct{}, cfg.MaxPending),
		cfg: cfg,
	}
}

// Acquire blocks until a slot is available or the context is cancelled.
func (c *Controller) Acquire(ctx context.Context) error {
	timeout := time.NewTimer(c.cfg.AcquireTimeout)
	defer timeout.Stop()
	select {
	case c.sem <- struct{}{}:
		c.mu.Lock()
		c.pending++
		c.mu.Unlock()
		return nil
	case <-timeout.C:
		return ErrBackpressure
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release frees a previously acquired slot.
func (c *Controller) Release() {
	select {
	case <-c.sem:
		c.mu.Lock()
		if c.pending > 0 {
			c.pending--
		}
		c.mu.Unlock()
	default:
	}
}

// Pending returns the current number of in-flight events.
func (c *Controller) Pending() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.pending
}
