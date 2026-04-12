// Package normalizer provides field-level normalization for WAL events
// before they are forwarded to downstream sinks.
package normalizer

import (
	"strings"

	"github.com/pgstream/pgstream/internal/wal"
)

// Rule defines a normalization rule for a specific table and column.
type Rule struct {
	Table  string
	Column string
	Mode   string // "lowercase", "uppercase", "trim", "trimspace"
}

// Config holds the normalizer configuration.
type Config struct {
	Rules []Rule
}

// Normalizer applies field normalization rules to WAL events.
type Normalizer struct {
	rules []Rule
}

// New creates a new Normalizer from the given config.
// If config is nil or has no rules, the normalizer is a no-op.
func New(cfg *Config) *Normalizer {
	if cfg == nil {
		return &Normalizer{}
	}
	return &Normalizer{rules: cfg.Rules}
}

// Apply normalizes fields in the event according to configured rules.
// Returns nil if the event is nil. Never mutates the original event.
func (n *Normalizer) Apply(event *wal.Event) *wal.Event {
	if event == nil {
		return nil
	}
	if len(n.rules) == 0 {
		return event
	}

	cloned := event.Clone()
	for _, rule := range n.rules {
		if !tableMatches(rule.Table, cloned.Table) {
			continue
		}
		if v, ok := cloned.Data[rule.Column]; ok {
			if s, ok := v.(string); ok {
				cloned.Data[rule.Column] = applyMode(s, rule.Mode)
			}
		}
	}
	return cloned
}

func tableMatches(pattern, table string) bool {
	if pattern == "*" || pattern == "" {
		return true
	}
	return strings.EqualFold(pattern, table)
}

func applyMode(s, mode string) string {
	switch mode {
	case "lowercase":
		return strings.ToLower(s)
	case "uppercase":
		return strings.ToUpper(s)
	case "trim":
		return strings.TrimSpace(s)
	case "trimspace":
		return strings.TrimSpace(s)
	default:
		return s
	}
}
