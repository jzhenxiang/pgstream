// Package masker provides field-level data masking for WAL events before
// they are forwarded to downstream sinks. Masking rules are defined per
// table and column and support several strategies: redact (replace with
// a fixed string), hash (SHA-256 hex digest), and partial (keep a
// configurable prefix and replace the remainder with '*').
package masker

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/your-org/pgstream/internal/wal"
)

// Strategy defines how a matched field value is masked.
type Strategy string

const (
	StrategyRedact  Strategy = "redact"
	StrategyHash    Strategy = "hash"
	StrategyPartial Strategy = "partial"

	defaultRedactValue  = "[REDACTED]"
	defaultPartialKeep  = 2
)

// Rule describes a masking rule for a single column.
type Rule struct {
	Table    string   // schema.table or just table
	Column   string
	Strategy Strategy
	// PartialKeep is the number of leading characters preserved when
	// Strategy == StrategyPartial. Defaults to 2 when zero.
	PartialKeep int
	// RedactValue overrides the default redaction string.
	RedactValue string
}

// Config holds all masking rules.
type Config struct {
	Rules []Rule
}

// Masker applies masking rules to WAL events.
type Masker struct {
	// index: "table.column" -> Rule
	rules map[string]Rule
}

// New creates a Masker from the supplied Config.
func New(cfg Config) *Masker {
	m := &Masker{rules: make(map[string]Rule, len(cfg.Rules))}
	for _, r := range cfg.Rules {
		key := normalize(r.Table) + "." + strings.ToLower(r.Column)
		m.rules[key] = r
	}
	return m
}

// Apply returns a shallow copy of the event with masked field values.
// If no rules match the event, the original pointer is returned unchanged.
func (m *Masker) Apply(event *wal.Event) *wal.Event {
	if event == nil || len(m.rules) == 0 {
		return event
	}
	tableKey := normalize(event.Table)
	masked := false
	copy := *event
	newFields := make(map[string]interface{}, len(event.Fields))
	for k, v := range event.Fields {
		key := tableKey + "." + strings.ToLower(k)
		if rule, ok := m.rules[key]; ok {
			newFields[k] = applyStrategy(rule, v)
			masked = true
		} else {
			newFields[k] = v
		}
	}
	if !masked {
		return event
	}
	copy.Fields = newFields
	return &copy
}

func applyStrategy(r Rule, v interface{}) interface{} {
	switch r.Strategy {
	case StrategyHash:
		sum := sha256.Sum256([]byte(fmt.Sprintf("%v", v)))
		return fmt.Sprintf("%x", sum)
	case StrategyPartial:
		s := fmt.Sprintf("%v", v)
		keep := r.PartialKeep
		if keep <= 0 {
			keep = defaultPartialKeep
		}
		if len(s) <= keep {
			return s
		}
		return s[:keep] + strings.Repeat("*", len(s)-keep)
	default: // StrategyRedact
		if r.RedactValue != "" {
			return r.RedactValue
		}
		return defaultRedactValue
	}
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
