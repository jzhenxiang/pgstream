// Package dedup provides a lightweight deduplication layer for WAL events.
// It tracks recently seen event keys using a fixed-size LRU-style window and
// allows callers to skip duplicate events that may arise from replication
// restarts or at-least-once delivery guarantees.
package dedup

import (
	"sync"
	"time"
)

// DefaultWindowSize is the number of event keys retained in memory.
const DefaultWindowSize = 1024

// Config holds tunable parameters for the deduplicator.
type Config struct {
	// WindowSize is the maximum number of keys to track. Defaults to DefaultWindowSize.
	WindowSize int
	// TTL is how long a key is considered "seen". Zero means no expiry.
	TTL time.Duration
}

type entry struct {
	seenAt time.Time
}

// Dedup tracks seen event keys and reports duplicates.
type Dedup struct {
	mu      sync.Mutex
	keys    map[string]entry
	order   []string
	window  int
	ttl     time.Duration
}

// New creates a new Dedup with the given config.
// If cfg is nil, defaults are applied.
func New(cfg *Config) *Dedup {
	win := DefaultWindowSize
	var ttl time.Duration
	if cfg != nil {
		if cfg.WindowSize > 0 {
			win = cfg.WindowSize
		}
		ttl = cfg.TTL
	}
	return &Dedup{
		keys:   make(map[string]entry, win),
		order:  make([]string, 0, win),
		window: win,
		ttl:    ttl,
	}
}

// IsDuplicate returns true if key has been seen within the active window/TTL.
// If the key is new (or expired), it is recorded and false is returned.
func (d *Dedup) IsDuplicate(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()

	if e, ok := d.keys[key]; ok {
		if d.ttl == 0 || now.Sub(e.seenAt) < d.ttl {
			return true
		}
		// expired — treat as new
	}

	// evict oldest entry if at capacity
	if len(d.order) >= d.window {
		oldest := d.order[0]
		d.order = d.order[1:]
		delete(d.keys, oldest)
	}

	d.keys[key] = entry{seenAt: now}
	d.order = append(d.order, key)
	return false
}

// Len returns the number of keys currently tracked.
func (d *Dedup) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.keys)
}

// Reset clears all tracked keys.
func (d *Dedup) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.keys = make(map[string]entry, d.window)
	d.order = d.order[:0]
}
