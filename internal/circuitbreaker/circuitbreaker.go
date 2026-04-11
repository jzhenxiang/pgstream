// Package circuitbreaker provides a simple circuit breaker to prevent
// cascading failures when a downstream sink becomes unavailable.
//
// The circuit moves through three states:
//   - Closed   – normal operation, requests pass through.
//   - Open     – too many consecutive failures; requests are rejected immediately.
//   - HalfOpen – a probe request is allowed to test recovery.
package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// ErrOpen is returned when the circuit is open and requests are blocked.
var ErrOpen = errors.New("circuit breaker is open")

// State represents the current state of the circuit breaker.
type State int

const (
	StateClosed   State = iota
	StateOpen
	StateHalfOpen
)

// Config holds tunable parameters for the circuit breaker.
type Config struct {
	// MaxFailures is the number of consecutive failures before opening.
	MaxFailures int
	// ResetTimeout is how long to wait in the Open state before moving to HalfOpen.
	ResetTimeout time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		MaxFailures:  5,
		ResetTimeout: 30 * time.Second,
	}
}

// CircuitBreaker guards calls to a downstream dependency.
type CircuitBreaker struct {
	mu           sync.Mutex
	cfg          Config
	state        State
	failures     int
	lastFailure  time.Time
}

// New creates a CircuitBreaker using the provided Config.
// Zero values in cfg are replaced with defaults.
func New(cfg Config) *CircuitBreaker {
	def := DefaultConfig()
	if cfg.MaxFailures <= 0 {
		cfg.MaxFailures = def.MaxFailures
	}
	if cfg.ResetTimeout <= 0 {
		cfg.ResetTimeout = def.ResetTimeout
	}
	return &CircuitBreaker{cfg: cfg}
}

// Allow reports whether the call should be attempted.
// It returns ErrOpen when the circuit is open and the reset timeout has not elapsed.
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return nil
	case StateOpen:
		if time.Since(cb.lastFailure) >= cb.cfg.ResetTimeout {
			cb.state = StateHalfOpen
			return nil
		}
		return ErrOpen
	case StateHalfOpen:
		return nil
	}
	return nil
}

// RecordSuccess resets the failure counter and closes the circuit.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = StateClosed
}

// RecordFailure increments the failure counter and opens the circuit when the
// threshold is reached.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= cb.cfg.MaxFailures {
		cb.state = StateOpen
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
