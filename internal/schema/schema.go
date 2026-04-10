// Package schema provides utilities for tracking and caching Postgres table
// schema information (column names and types) discovered during WAL streaming.
package schema

import (
	"fmt"
	"sync"
)

// Column holds metadata about a single table column.
type Column struct {
	Name     string
	Type     string
	Position int
}

// TableSchema holds the ordered column definitions for a table.
type TableSchema struct {
	Schema  string
	Table   string
	Columns []Column
}

// Key returns a fully-qualified "schema.table" identifier.
func (t *TableSchema) Key() string {
	return fmt.Sprintf("%s.%s", t.Schema, t.Table)
}

// Cache stores table schemas in memory and is safe for concurrent use.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]*TableSchema
}

// New returns an initialised, empty Cache.
func New() *Cache {
	return &Cache{
		entries: make(map[string]*TableSchema),
	}
}

// Set stores or replaces the schema for the given table.
func (c *Cache) Set(ts *TableSchema) {
	if ts == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[ts.Key()] = ts
}

// Get retrieves the schema for "schema.table". The second return value
// indicates whether the entry was found.
func (c *Cache) Get(schema, table string) (*TableSchema, bool) {
	key := fmt.Sprintf("%s.%s", schema, table)
	c.mu.RLock()
	defer c.mu.RUnlock()
	ts, ok := c.entries[key]
	return ts, ok
}

// Delete removes the schema entry for the given table, if present.
func (c *Cache) Delete(schema, table string) {
	key := fmt.Sprintf("%s.%s", schema, table)
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// Len returns the number of cached table schemas.
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}
