// Package tee provides a sink that duplicates events to multiple sinks
// while continuing even if one sink fails, collecting all errors.
package tee

import (
	"context"
	"errors"
	"fmt"

	"pgstream/internal/sink"
	"pgstream/internal/wal"
)

// Tee fans out events to all registered sinks, collecting errors from each.
// Unlike fanout, Tee does not abort on the first error — it attempts all sinks
// and returns a combined error if any failed.
type Tee struct {
	sinks []sink.Sink
}

// New returns a Tee that writes to all provided sinks.
// At least one sink must be provided.
func New(sinks ...sink.Sink) (*Tee, error) {
	if len(sinks) == 0 {
		return nil, errors.New("tee: at least one sink is required")
	}
	for i, s := range sinks {
		if s == nil {
			return nil, fmt.Errorf("tee: sink at index %d is nil", i)
		}
	}
	return &Tee{sinks: sinks}, nil
}

// Send writes the event to every sink. All sinks are attempted regardless of
// individual failures. A combined error is returned if one or more sinks fail.
func (t *Tee) Send(ctx context.Context, event *wal.Event) error {
	var errs []error
	for _, s := range t.sinks {
		if err := s.Send(ctx, event); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// Len returns the number of sinks registered with this Tee.
func (t *Tee) Len() int {
	return len(t.sinks)
}
