// Package transform provides event transformation utilities for WAL messages
// before they are forwarded to a sink. Transformations include field renaming,
// filtering out sensitive columns, and adding metadata fields.
package transform

import (
	"fmt"
	"time"
)

// Event represents a decoded WAL event ready for transformation.
type Event map[string]any

// Config holds the transformation configuration.
type Config struct {
	// RedactColumns is a list of column names whose values will be replaced with "[REDACTED]".
	RedactColumns []string `yaml:"redact_columns"`
	// AddMetadata controls whether pgstream metadata fields are injected.
	AddMetadata bool `yaml:"add_metadata"`
	// RenameColumns maps original column names to new names.
	RenameColumns map[string]string `yaml:"rename_columns"`
}

// Transformer applies a set of transformations to WAL events.
type Transformer struct {
	cfg    Config
	redact map[string]struct{}
}

// New creates a new Transformer from the given Config.
func New(cfg Config) *Transformer {
	redact := make(map[string]struct{}, len(cfg.RedactColumns))
	for _, col := range cfg.RedactColumns {
		redact[col] = struct{}{}
	}
	return &Transformer{cfg: cfg, redact: redact}
}

// Apply runs all configured transformations on the event and returns the result.
// The original event is not modified; a shallow copy is returned.
func (t *Transformer) Apply(table string, event Event) (Event, error) {
	if event == nil {
		return nil, fmt.Errorf("transform: nil event for table %q", table)
	}

	out := make(Event, len(event))
	for k, v := range event {
		out[k] = v
	}

	// Redact sensitive columns.
	for col := range t.redact {
		if _, ok := out[col]; ok {
			out[col] = "[REDACTED]"
		}
	}

	// Rename columns.
	for src, dst := range t.cfg.RenameColumns {
		if val, ok := out[src]; ok {
			out[dst] = val
			delete(out, src)
		}
	}

	// Inject metadata.
	if t.cfg.AddMetadata {
		out["_pgstream_table"] = table
		out["_pgstream_ts"] = time.Now().UTC().Format(time.RFC3339Nano)
	}

	return out, nil
}
