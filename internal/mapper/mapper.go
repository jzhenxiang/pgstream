// Package mapper provides field-level key remapping for WAL events.
// It allows renaming top-level data fields before forwarding to a sink,
// useful for schema normalisation across heterogeneous consumers.
package mapper

import (
	"errors"

	"github.com/pgstream/pgstream/internal/wal"
)

// Config holds a per-table mapping of old field names to new field names.
type Config struct {
	// Rules maps table names to a map of {oldField: newField}.
	Rules map[string]map[string]string
}

// Mapper renames fields in WAL events according to configured rules.
type Mapper struct {
	rules map[string]map[string]string
}

// New creates a Mapper from the given Config.
// Returns an error if any mapping entry has a blank source or target key.
func New(cfg Config) (*Mapper, error) {
	for table, fields := range cfg.Rules {
		for src, dst := range fields {
			if src == "" || dst == "" {
				return nil, errors.New("mapper: table " + table + " has blank field mapping")
			}
		}
	}
	return &Mapper{rules: cfg.Rules}, nil
}

// Apply returns a shallow-copied event with fields renamed according to the
// configured rules for the event's table. If no rule matches the event, the
// original pointer is returned unchanged.
func (m *Mapper) Apply(event *wal.Event) *wal.Event {
	if event == nil {
		return nil
	}
	mapping, ok := m.rules[event.Table]
	if !ok || len(mapping) == 0 {
		return event
	}
	out := *event
	renamed := make(map[string]any, len(event.Data))
	for k, v := range event.Data {
		if newKey, found := mapping[k]; found {
			renamed[newKey] = v
		} else {
			renamed[k] = v
		}
	}
	out.Data = renamed
	return &out
}
