// Package nackhandler provides a negative-acknowledgement handler that routes
// failed events to a configurable fallback sink (e.g. a DLQ or a retry queue)
// after a maximum number of delivery attempts has been exceeded.
package nackhandler

import (
	"context"
	"errors"
	"fmt"

	"github.com/your-org/pgstream/internal/sink"
	"github.com/your-org/pgstream/internal/wal"
)

// DefaultMaxAttempts is used when Config.MaxAttempts is zero.
const DefaultMaxAttempts = 3

// Config holds options for the Handler.
type Config struct {
	// MaxAttempts is the number of delivery attempts before an event is
	// forwarded to the Fallback sink. Zero uses DefaultMaxAttempts.
	MaxAttempts int

	// Fallback receives events that have exhausted their attempts.
	Fallback sink.Sink
}

func (c *Config) validate() error {
	if c == nil {
		return errors.New("nackhandler: nil config")
	}
	if c.Fallback == nil {
		return errors.New("nackhandler: fallback sink is required")
	}
	return nil
}

// Handler wraps a primary sink and intercepts send errors, forwarding
// exhausted events to the configured fallback sink.
type Handler struct {
	maxAttempts int
	fallback    sink.Sink
	attempts    map[string]int
}

// New creates a new Handler from cfg.
func New(cfg Config) (*Handler, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	max := cfg.MaxAttempts
	if max <= 0 {
		max = DefaultMaxAttempts
	}
	return &Handler{
		maxAttempts: max,
		fallback:    cfg.Fallback,
		attempts:    make(map[string]int),
	}, nil
}

// Handle is called when a primary sink returns an error for event. It
// increments the attempt counter and, once MaxAttempts is reached, forwards
// the event to the fallback sink and clears the counter.
func (h *Handler) Handle(ctx context.Context, event *wal.Event, sendErr error) error {
	if event == nil {
		return errors.New("nackhandler: nil event")
	}
	key := eventKey(event)
	h.attempts[key]++
	if h.attempts[key] < h.maxAttempts {
		return sendErr // caller may retry
	}
	delete(h.attempts, key)
	if err := h.fallback.Send(ctx, event); err != nil {
		return fmt.Errorf("nackhandler: fallback send failed: %w", err)
	}
	return nil
}

// Attempts returns the current attempt count for an event key (testing helper).
func (h *Handler) Attempts(key string) int {
	return h.attempts[key]
}

func eventKey(e *wal.Event) string {
	return fmt.Sprintf("%s.%s@%s", e.Schema, e.Table, e.LSN)
}
