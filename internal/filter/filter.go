// Package filter provides table-level filtering for WAL events.
package filter

import "strings"

// Filter determines whether WAL events for a given table should be processed.
type Filter struct {
	allowList map[string]struct{}
	denyList  map[string]struct{}
}

// Config holds the filter configuration.
type Config struct {
	// AllowTables is a list of "schema.table" patterns to include.
	// If non-empty, only these tables are processed.
	AllowTables []string
	// DenyTables is a list of "schema.table" patterns to exclude.
	DenyTables []string
}

// New creates a new Filter from the given config.
func New(cfg Config) *Filter {
	f := &Filter{
		allowList: make(map[string]struct{}, len(cfg.AllowTables)),
		denyList:  make(map[string]struct{}, len(cfg.DenyTables)),
	}
	for _, t := range cfg.AllowTables {
		f.allowList[normalize(t)] = struct{}{}
	}
	for _, t := range cfg.DenyTables {
		f.denyList[normalize(t)] = struct{}{}
	}
	return f
}

// Allow returns true if the given schema.table should be processed.
func (f *Filter) Allow(schema, table string) bool {
	key := normalize(schema + "." + table)

	if _, denied := f.denyList[key]; denied {
		return false
	}
	if len(f.allowList) == 0 {
		return true
	}
	_, allowed := f.allowList[key]
	return allowed
}

// IsEmpty returns true when no filtering rules are configured.
func (f *Filter) IsEmpty() bool {
	return len(f.allowList) == 0 && len(f.denyList) == 0
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
