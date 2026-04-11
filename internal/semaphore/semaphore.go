// Package semaphore provides a weighted semaphore for controlling
// concurrent access to shared resources within the pgstream pipeline.
package semaphore

import (
	"context"
	"errors"
	"sync"
)

// ErrAcquire is returned when the semaphore cannot be acquired.
var ErrAcquire = errors.New("semaphore: failed to acquire")

// Semaphore is a counting semaphore backed by a buffered channel.
type Semaphore struct {
	mu      sync.Mutex
	ch      chan struct{}
	max     int
	current int
}

// New creates a new Semaphore with the given maximum concurrency.
// Returns an error if n is less than 1.
func New(n int) (*Semaphore, error) {
	if n < 1 {
		return nil, errors.New("semaphore: max must be at least 1")
	}
	return &Semaphore{
		ch:  make(chan struct{}, n),
		max: n,
	}, nil
}

// Acquire blocks until a slot is available or the context is cancelled.
func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.ch <- struct{}{}:
		s.mu.Lock()
		s.current++
		s.mu.Unlock()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TryAcquire attempts to acquire a slot without blocking.
// Returns ErrAcquire if no slot is available.
func (s *Semaphore) TryAcquire() error {
	select {
	case s.ch <- struct{}{}:
		s.mu.Lock()
		s.current++
		s.mu.Unlock()
		return nil
	default:
		return ErrAcquire
	}
}

// Release frees one slot in the semaphore.
func (s *Semaphore) Release() {
	select {
	case <-s.ch:
		s.mu.Lock()
		s.current--
		s.mu.Unlock()
	default:
	}
}

// Current returns the number of currently held slots.
func (s *Semaphore) Current() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.current
}

// Max returns the maximum concurrency of the semaphore.
func (s *Semaphore) Max() int {
	return s.max
}

// Available returns the number of slots that can still be acquired
// without blocking.
func (s *Semaphore) Available() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.max - s.current
}
