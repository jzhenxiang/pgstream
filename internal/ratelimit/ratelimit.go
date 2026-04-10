// Package ratelimit provides a token-bucket rate limiter for controlling
// the throughput of WAL events through the pipeline.
package ratelimit

import (
	"context"
	"fmt"
	"time"
)

// Config holds rate limiter configuration.
type Config struct {
	// EventsPerSecond is the maximum number of events allowed per second.
	// A value of 0 disables rate limiting.
	EventsPerSecond int
}

// Limiter controls the rate at which events are processed.
type Limiter struct {
	tokens  chan struct{}
	stop    chan struct{}
	disabled bool
}

// New creates a new Limiter from the given Config.
// If EventsPerSecond is 0, the limiter is disabled and Wait always returns nil.
func New(cfg Config) (*Limiter, error) {
	if cfg.EventsPerSecond < 0 {
		return nil, fmt.Errorf("ratelimit: EventsPerSecond must be >= 0, got %d", cfg.EventsPerSecond)
	}

	l := &Limiter{
		stop: make(chan struct{}),
	}

	if cfg.EventsPerSecond == 0 {
		l.disabled = true
		return l, nil
	}

	l.tokens = make(chan struct{}, cfg.EventsPerSecond)

	// Pre-fill bucket.
	for i := 0; i < cfg.EventsPerSecond; i++ {
		l.tokens <- struct{}{}
	}

	interval := time.Second / time.Duration(cfg.EventsPerSecond)
	go l.refill(interval)

	return l, nil
}

// Wait blocks until a token is available or the context is cancelled.
// Returns nil immediately if the limiter is disabled.
func (l *Limiter) Wait(ctx context.Context) error {
	if l.disabled {
		return nil
	}
	select {
	case <-l.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-l.stop:
		return fmt.Errorf("ratelimit: limiter stopped")
	}
}

// Stop shuts down the background refill goroutine.
func (l *Limiter) Stop() {
	if !l.disabled {
		close(l.stop)
	}
}

func (l *Limiter) refill(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			select {
			case l.tokens <- struct{}{}:
			default:
				// Bucket full; drop the token.
			}
		case <-l.stop:
			return
		}
	}
}
