// Package matcher provides pattern-based table and column matching
// for use in filtering, transformation, and routing pipelines.
package matcher

import (
	"errors"
	"path"
	"strings"
)

// Matcher evaluates whether a table or column name matches a set of patterns.
type Matcher struct {
	patterns []string
}

// New returns a Matcher configured with the given patterns.
// Patterns support glob-style wildcards (e.g. "public.*", "*.users").
// Returns an error if no patterns are provided.
func New(patterns []string) (*Matcher, error) {
	if len(patterns) == 0 {
		return nil, errors.New("matcher: at least one pattern is required")
	}
	norm := make([]string, 0, len(patterns))
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			return nil, errors.New("matcher: empty pattern is not allowed")
		}
		norm = append(norm, strings.ToLower(p))
	}
	return &Matcher{patterns: norm}, nil
}

// Match reports whether the given value matches any of the configured patterns.
// Matching is case-insensitive and supports glob wildcards.
func (m *Matcher) Match(value string) bool {
	v := strings.ToLower(strings.TrimSpace(value))
	for _, p := range m.patterns {
		if ok, _ := path.Match(p, v); ok {
			return true
		}
	}
	return false
}

// MatchAny reports whether any of the given values matches at least one pattern.
func (m *Matcher) MatchAny(values []string) bool {
	for _, v := range values {
		if m.Match(v) {
			return true
		}
	}
	return false
}

// Patterns returns a copy of the configured patterns.
func (m *Matcher) Patterns() []string {
	out := make([]string, len(m.patterns))
	copy(out, m.patterns)
	return out
}
