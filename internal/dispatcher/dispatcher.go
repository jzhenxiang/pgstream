// Package dispatcher routes WAL events to one or more sinks based on
// table-level routing rules. It wraps the fanout sink and applies per-table
// sink mappings before forwarding events.
package dispatcher

import (
	"context"
	"fmt"

	"github.com/pgstream/pgstream/internal/sink"
	"github.com/pgstream/pgstream/internal/wal"
)

// Route maps a table name (schema.table) to a set of sinks.
type Route struct {
	Table string
	Sinks []sink.Sink
}

// Dispatcher routes events to sinks according to routing rules.
// Events whose table does not match any rule are forwarded to the
// default sink when one is configured.
type Dispatcher struct {
	routes       map[string][]sink.Sink
	defaultSinks []sink.Sink
}

// New creates a Dispatcher from the provided routes and an optional
// default sink list. Returns an error when no routes and no default
// sinks are provided.
func New(routes []Route, defaultSinks []sink.Sink) (*Dispatcher, error) {
	if len(routes) == 0 && len(defaultSinks) == 0 {
		return nil, fmt.Errorf("dispatcher: at least one route or default sink must be provided")
	}

	table := make(map[string][]sink.Sink, len(routes))
	for _, r := range routes {
		if r.Table == "" {
			return nil, fmt.Errorf("dispatcher: route table name must not be empty")
		}
		if len(r.Sinks) == 0 {
			return nil, fmt.Errorf("dispatcher: route %q must have at least one sink", r.Table)
		}
		table[r.Table] = r.Sinks
	}

	return &Dispatcher{
		routes:       table,
		defaultSinks: defaultSinks,
	}, nil
}

// Dispatch sends the event to all sinks registered for the event's table.
// When no matching route exists the event is forwarded to the default sinks.
// Returns an error if any sink returns an error.
func (d *Dispatcher) Dispatch(ctx context.Context, event *wal.Event) error {
	targets := d.sinksFor(event)
	for _, s := range targets {
		if err := s.Send(ctx, event); err != nil {
			return fmt.Errorf("dispatcher: send to sink: %w", err)
		}
	}
	return nil
}

func (d *Dispatcher) sinksFor(event *wal.Event) []sink.Sink {
	if event == nil {
		return d.defaultSinks
	}
	key := event.Table
	if sinks, ok := d.routes[key]; ok {
		return sinks
	}
	return d.defaultSinks
}
