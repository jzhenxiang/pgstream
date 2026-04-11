// Package throttle provides an adaptive throughput throttler that slows
// event processing when downstream sinks signal backpressure.
package throttle

import (
	"context"
	"errors"
	"sync"
	"time"
)

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		MinDelay: 0,
		MaxDelay: 2 * time.Second,
		Step:     100 * time.Millisecond,
	}
}

// Config controls throttle behaviour.
type Config struct {
	// MinDelay is the lowest possible inter-event delay (0 means no delay).
	MinDelay time.Duration
	// MaxDelay caps the delay so the pipeline never fully stalls.
	MaxDelay time.Duration
	// Step is the amount added/subtracted on each backpressure signal.
	Step time.Duration
}

// Throttle adaptively delays event processing.
type Throttle struct {
	cfg     Config
	current time.Duration
	mu      sync.Mutex
}

// New creates a Throttle from cfg. Zero-value fields are replaced with defaults.
func New(cfg Config) (*Throttle, error) {
	def := DefaultConfig()
	if cfg.MaxDelay == 0 {
		cfg.MaxDelay = def.MaxDelay
	}
	if cfg.Step == 0 {
		cfg.Step = def.Step
	}
	if cfg.MaxDelay < cfg.MinDelay {
		return nil, errors.New("throttle: MaxDelay must be >= MinDelay")
	}
	return &Throttle{cfg: cfg, current: cfg.MinDelay}, nil
}

// Increase raises the current delay by one step, up to MaxDelay.
func (t *Throttle) Increase() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.current += t.cfg.Step
	if t.current > t.cfg.MaxDelay {
		t.current = t.cfg.MaxDelay
	}
}

// Decrease lowers the current delay by one step, down to MinDelay.
func (t *Throttle) Decrease() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.current -= t.cfg.Step
	if t.current < t.cfg.MinDelay {
		t.current = t.cfg.MinDelay
	}
}

// Wait blocks for the current delay or until ctx is cancelled.
func (t *Throttle) Wait(ctx context.Context) error {
	t.mu.Lock()
	d := t.current
	t.mu.Unlock()
	if d == 0 {
		return nil
	}
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Current returns the active delay.
func (t *Throttle) Current() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.current
}
