// Package redactor provides field-level redaction for WAL events before
// they are forwarded to downstream sinks. Rules are matched by table name
// and a configurable strategy (blank, hash, or partial) is applied.
package redactor

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/pgstream/pgstream/internal/wal"
)

// Strategy controls how a matched field value is redacted.
type Strategy string

const (
	StrategyBlank   Strategy = "blank"
	StrategyHash    Strategy = "hash"
	StrategyPartial Strategy = "partial"
)

// Rule describes which columns in which tables should be redacted.
type Rule struct {
	Table    string
	Columns  []string
	Strategy Strategy
}

// Config holds the list of redaction rules.
type Config struct {
	Rules []Rule
}

// Redactor applies field-level redaction to WAL events.
type Redactor struct {
	cfg Config
}

// New returns a Redactor configured with the provided rules.
// An empty config is valid and results in a no-op redactor.
func New(cfg Config) *Redactor {
	return &Redactor{cfg: cfg}
}

// Apply returns a copy of the event with configured fields redacted.
// If event is nil, nil is returned.
func (r *Redactor) Apply(event *wal.Event) *wal.Event {
	if event == nil {
		return nil
	}
	if len(r.cfg.Rules) == 0 {
		return event
	}

	cloned := event.Clone()
	for _, rule := range r.cfg.Rules {
		if !tableMatches(rule.Table, cloned.Table) {
			continue
		}
		for _, col := range rule.Columns {
			if v, ok := cloned.Data[col]; ok {
				cloned.Data[col] = redact(fmt.Sprintf("%v", v), rule.Strategy)
			}
		}
	}
	return cloned
}

func tableMatches(pattern, table string) bool {
	return strings.EqualFold(pattern, table) || pattern == "*"
}

func redact(value string, s Strategy) string {
	switch s {
	case StrategyHash:
		sum := sha256.Sum256([]byte(value))
		return fmt.Sprintf("%x", sum[:8])
	case StrategyPartial:
		if len(value) <= 2 {
			return "***"
		}
		return string(value[0]) + strings.Repeat("*", len(value)-2) + string(value[len(value)-1])
	default: // StrategyBlank
		return ""
	}
}
