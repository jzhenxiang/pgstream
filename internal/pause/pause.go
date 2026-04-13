// Package pause provides a pausable controller that allows the pipeline
// to be temporarily halted and resumed, e.g. during backpressure events
// or operator-triggered maintenance windows.
package pause

import (
	"context"
	"sync"
)

// Controller manages the paused/running state of a pipeline component.
type Controller struct {
	mu     sync.RWMutex
	paused bool
	resume chan struct{}
}

// New returns a new Controller in the running (unpaused) state.
func New() *Controller {
	return &Controller{
		resume: make(chan struct{}),
	}
}

// Pause transitions the controller into the paused state.
// Calling Pause when already paused is a no-op.
func (c *Controller) Pause() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.paused {
		c.paused = true
		c.resume = make(chan struct{})
	}
}

// Resume transitions the controller back to the running state.
// Calling Resume when not paused is a no-op.
func (c *Controller) Resume() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.paused {
		c.paused = false
		close(c.resume)
	}
}

// IsPaused reports whether the controller is currently paused.
func (c *Controller) IsPaused() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.paused
}

// Wait blocks until the controller is in the running state or ctx is done.
// Returns ctx.Err() if the context is cancelled while waiting.
func (c *Controller) Wait(ctx context.Context) error {
	c.mu.RLock()
	if !c.paused {
		c.mu.RUnlock()
		return nil
	}
	ch := c.resume
	c.mu.RUnlock()

	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// PauseAndWait transitions the controller into the paused state and then
// blocks until Resume is called or ctx is done. This is a convenience
// method combining Pause and Wait for callers that own both operations.
func (c *Controller) PauseAndWait(ctx context.Context) error {
	c.Pause()
	return c.Wait(ctx)
}
