// Package selector provides column-level field selection for WAL events,
// allowing only a configured subset of columns to be forwarded downstream.
package selector

import (
	"errors"

	"pgstream/internal/wal"
)

// Config holds the per-table column allow-lists.
type Config struct {
	// Rules maps table names (schema.table or just table) to the set of
	// column names that should be retained. All other columns are dropped.
	Rules map[string][]string
}

// Selector filters event columns according to a configured allow-list.
type Selector struct {
	rules map[string]map[string]struct{}
}

// New creates a Selector from cfg. An empty config is valid and produces a
// no-op selector.
func New(cfg Config) (*Selector, error) {
	if cfg.Rules == nil {
		return &Selector{rules: make(map[string]map[string]struct{})}, nil
	}

	rules := make(map[string]map[string]struct{}, len(cfg.Rules))
	for table, cols := range cfg.Rules {
		if table == "" {
			return nil, errors.New("selector: table name must not be empty")
		}
		set := make(map[string]struct{}, len(cols))
		for _, c := range cols {
			set[c] = struct{}{}
		}
		rules[table] = set
	}
	return &Selector{rules: rules}, nil
}

// Apply returns a shallow copy of event with only the allowed columns retained.
// If no rule exists for the event's table the event is returned unchanged.
// A nil event is returned as-is.
func (s *Selector) Apply(event *wal.Event) *wal.Event {
	if event == nil {
		return nil
	}

	allowed, ok := s.rules[event.Table]
	if !ok {
		allowed, ok = s.rules[event.Schema+"."+event.Table]
	}
	if !ok || len(allowed) == 0 {
		return event
	}

	filtered := make([]wal.Column, 0, len(allowed))
	for _, col := range event.Columns {
		if _, keep := allowed[col.Name]; keep {
			filtered = append(filtered, col)
		}
	}

	copy := *event
	copy.Columns = filtered
	return &copy
}
