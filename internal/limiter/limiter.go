package limiter

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrLimitExceeded is returned when the event rate limit is exceeded and the
// caller has not provided enough budget to wait.
var ErrLimitExceeded = errors.New("limiter: rate limit exceeded")

// Config holds configuration for the event limiter.
type Config struct {
	// MaxEventsPerSecond is the maximum number of events allowed per second.
	// A value of zero disables limiting.
	MaxEventsPerSecond int
	// BurstSize is the number of events that may be processed in a burst.
	// Defaults to MaxEventsPerSecond when zero.
	BurstSize int
}

// Limiter enforces a maximum event throughput using a token-bucket approach.
type Limiter struct {
	cfg    Config
	mu     sync.Mutex
	tokens float64
	lastAt time.Time
}

// New creates a new Limiter from cfg. If MaxEventsPerSecond is zero the limiter
// is a no-op and Allow always returns nil.
func New(cfg Config) (*Limiter, error) {
	if cfg.MaxEventsPerSecond < 0 {
		return nil, errors.New("limiter: MaxEventsPerSecond must be >= 0")
	}
	if cfg.BurstSize == 0 && cfg.MaxEventsPerSecond > 0 {
		cfg.BurstSize = cfg.MaxEventsPerSecond
	}
	return &Limiter{
		cfg:    cfg,
		tokens: float64(cfg.BurstSize),
		lastAt: time.Now(),
	}, nil
}

// Allow checks whether one event may be processed right now. It refills tokens
// based on elapsed time and consumes one token. If no tokens are available it
// returns ErrLimitExceeded. When limiting is disabled it always returns nil.
func (l *Limiter) Allow(_ context.Context) error {
	if l.cfg.MaxEventsPerSecond == 0 {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastAt).Seconds()
	l.lastAt = now

	l.tokens += elapsed * float64(l.cfg.MaxEventsPerSecond)
	if max := float64(l.cfg.BurstSize); l.tokens > max {
		l.tokens = max
	}

	if l.tokens < 1 {
		return ErrLimitExceeded
	}
	l.tokens--
	return nil
}

// Reset restores the token bucket to its full burst capacity.
func (l *Limiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.tokens = float64(l.cfg.BurstSize)
	l.lastAt = time.Now()
}
