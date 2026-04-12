// Package validator provides schema-based validation for WAL events before
// they are forwarded to downstream sinks. Rules are matched by table name and
// each rule can require specific columns to be present and non-nil.
package validator

import (
	"errors"
	"fmt"

	"github.com/pgstream/pgstream/internal/wal"
)

// Rule describes validation constraints for a single table.
type Rule struct {
	// Table is the fully-qualified table name (schema.table) or just table.
	Table string
	// RequiredColumns lists column names that must be present and non-nil.
	RequiredColumns []string
}

// Config holds the set of validation rules.
type Config struct {
	Rules []Rule
}

// Validator applies column-presence rules to WAL events.
type Validator struct {
	// index maps normalised table name -> rule
	index map[string]Rule
}

// ErrValidation is returned when an event fails validation.
var ErrValidation = errors.New("validation failed")

// New creates a Validator from cfg. An empty config produces a no-op validator
// that accepts every event.
func New(cfg Config) *Validator {
	idx := make(map[string]Rule, len(cfg.Rules))
	for _, r := range cfg.Rules {
		idx[r.Table] = r
	}
	return &Validator{index: idx}
}

// Validate checks event against the configured rules. It returns nil when no
// rule matches (allow-all) or when all required columns are satisfied.
func (v *Validator) Validate(event *wal.Event) error {
	if event == nil {
		return nil
	}
	rule, ok := v.index[event.Table]
	if !ok {
		// no rule for this table – pass through
		return nil
	}
	for _, col := range rule.RequiredColumns {
		val, exists := event.Data[col]
		if !exists || val == nil {
			return fmt.Errorf("%w: table %q missing required column %q",
				ErrValidation, event.Table, col)
		}
	}
	return nil
}

// RuleCount returns the number of configured rules.
func (v *Validator) RuleCount() int {
	return len(v.index)
}
