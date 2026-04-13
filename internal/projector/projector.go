// Package projector applies a column projection to WAL events, retaining
// only the fields explicitly listed in the configuration. This is useful
// for reducing payload size before forwarding to a sink.
package projector

import (
	"errors"

	"github.com/your-org/pgstream/internal/wal"
)

// Config holds the per-table column allow-lists used during projection.
type Config struct {
	// Rules maps a table name (schema.table or just table) to the set of
	// column names that should be retained in the output event.
	Rules map[string][]string
}

// Projector filters event columns according to configured rules.
type Projector struct {
	cfg Config
}

// New returns a Projector for the given Config.
// An empty Config is valid; events will be returned unchanged.
func New(cfg Config) (*Projector, error) {
	for table, cols := range cfg.Rules {
		if table == "" {
			return nil, errors.New("projector: table name must not be blank")
		}
		if len(cols) == 0 {
			return nil, errors.New("projector: column list must not be empty for table " + table)
		}
	}
	return &Projector{cfg: cfg}, nil
}

// Apply returns a shallow copy of the event with only the allowed columns
// retained. If no rule matches the event's table the event is returned as-is.
// A nil event returns nil.
func (p *Projector) Apply(event *wal.Event) *wal.Event {
	if event == nil {
		return nil
	}

	allowed, ok := p.cfg.Rules[event.Table]
	if !ok {
		return event
	}

	set := make(map[string]struct{}, len(allowed))
	for _, c := range allowed {
		set[c] = struct{}{}
	}

	filtered := make(map[string]any, len(set))
	for k, v := range event.Data {
		if _, keep := set[k]; keep {
			filtered[k] = v
		}
	}

	copy := *event
	copy.Data = filtered
	return &copy
}
