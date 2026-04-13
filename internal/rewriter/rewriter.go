// Package rewriter provides field-level value rewriting for WAL events.
// Rules are matched by table and column name; matched values are replaced
// using a static mapping supplied at construction time.
package rewriter

import (
	"fmt"

	"pgstream/internal/wal"
)

// Rule describes a single rewrite operation for a table/column pair.
type Rule struct {
	Table   string            // exact table name (schema-qualified optional)
	Column  string            // column to rewrite
	Mapping map[string]string // old value -> new value
}

// Config holds the set of rewrite rules.
type Config struct {
	Rules []Rule
}

// Rewriter applies value rewrites to WAL events.
type Rewriter struct {
	cfg Config
}

// New returns a Rewriter for the given config.
// An empty config is valid and produces a no-op rewriter.
func New(cfg Config) *Rewriter {
	return &Rewriter{cfg: cfg}
}

// Apply rewrites fields in a copy of the event according to the configured
// rules. If event is nil, nil is returned. If no rules match, the original
// pointer is returned unchanged.
func (r *Rewriter) Apply(event *wal.Event) *wal.Event {
	if event == nil {
		return nil
	}
	if len(r.cfg.Rules) == 0 {
		return event
	}

	modified := false
	cloned := *event
	cloned.Fields = make(map[string]interface{}, len(event.Fields))
	for k, v := range event.Fields {
		cloned.Fields[k] = v
	}

	for _, rule := range r.cfg.Rules {
		if rule.Table != "" && rule.Table != event.Table {
			continue
		}
		raw, ok := cloned.Fields[rule.Column]
		if !ok {
			continue
		}
		strVal := fmt.Sprintf("%v", raw)
		if replacement, found := rule.Mapping[strVal]; found {
			cloned.Fields[rule.Column] = replacement
			modified = true
		}
	}

	if !modified {
		return event
	}
	return &cloned
}
