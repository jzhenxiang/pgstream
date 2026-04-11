// Package fanout provides a multi-sink dispatcher that forwards WAL events
// to multiple sinks concurrently, collecting all errors.
package fanout

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/pgstream/pgstream/internal/sink"
	"github.com/pgstream/pgstream/internal/wal"
)

// Fanout dispatches each WAL event to all registered sinks concurrently.
type Fanout struct {
	sinks []sink.Sink
}

// New creates a Fanout with the provided sinks.
// At least one sink must be provided.
func New(sinks ...sink.Sink) (*Fanout, error) {
	if len(sinks) == 0 {
		return nil, fmt.Errorf("fanout: at least one sink is required")
	}
	for i, s := range sinks {
		if s == nil {
			return nil, fmt.Errorf("fanout: sink at index %d is nil", i)
		}
	}
	return &Fanout{sinks: sinks}, nil
}

// Send dispatches the event to all sinks concurrently.
// If any sink returns an error, all errors are combined and returned.
func (f *Fanout) Send(ctx context.Context, event *wal.Event) error {
	var (
		mu   sync.Mutex
		errs []string
		wg   sync.WaitGroup
	)

	for _, s := range f.sinks {
		wg.Add(1)
		go func(s sink.Sink) {
			defer wg.Done()
			if err := s.Send(ctx, event); err != nil {
				mu.Lock()
				errs = append(errs, err.Error())
				mu.Unlock()
			}
		}(s)
	}

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("fanout: %d sink(s) failed: %s", len(errs), strings.Join(errs, "; "))
	}
	return nil
}

// Len returns the number of registered sinks.
func (f *Fanout) Len() int {
	return len(f.sinks)
}
