// Package coalesce merges multiple WAL events for the same row key
// within a configurable window, emitting only the latest version.
package coalesce

import (
	"errors"
	"sync"
	"time"

	"pgstream/internal/wal"
)

const (
	defaultWindowSize = 100
	defaultQuietPeriod = 50 * time.Millisecond
)

// Coalescer buffers events keyed by table+primary-key and flushes the
// latest event for each key after the quiet period elapses with no
// new arrivals for that key.
type Coalescer struct {
	mu          sync.Mutex
	window      map[string]*wal.Event
	order       []string
	maxSize     int
	quietPeriod time.Duration
	flush       func(*wal.Event) error
	timers      map[string]*time.Timer
}

// New returns a Coalescer that calls flush for each coalesced event.
// maxSize is the maximum number of distinct keys held before an
// immediate flush is forced. quietPeriod controls the debounce window.
func New(maxSize int, quietPeriod time.Duration, flush func(*wal.Event) error) (*Coalescer, error) {
	if flush == nil {
		return nil, errors.New("coalesce: flush func must not be nil")
	}
	if maxSize <= 0 {
		maxSize = defaultWindowSize
	}
	if quietPeriod <= 0 {
		quietPeriod = defaultQuietPeriod
	}
	return &Coalescer{
		window:      make(map[string]*wal.Event),
		timers:      make(map[string]*time.Timer),
		maxSize:     maxSize,
		quietPeriod: quietPeriod,
		flush:       flush,
	}, nil
}

// Add buffers the event. If the key already exists the previous event is
// replaced and the debounce timer is reset. When the buffer is full the
// oldest entry is flushed immediately.
func (c *Coalescer) Add(ev *wal.Event) error {
	if ev == nil {
		return nil
	}
	key := ev.Table + "|" + ev.RowID

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.window[key]; !exists {
		if len(c.window) >= c.maxSize {
			if err := c.evictOldest(); err != nil {
				return err
			}
		}
		c.order = append(c.order, key)
	}
	c.window[key] = ev

	if t, ok := c.timers[key]; ok {
		t.Reset(c.quietPeriod)
	} else {
		c.timers[key] = time.AfterFunc(c.quietPeriod, func() {
			c.mu.Lock()
			deferred := c.window[key]
			c.remove(key)
			c.mu.Unlock()
			if deferred != nil {
				_ = c.flush(deferred)
			}
		})
	}
	return nil
}

// evictOldest flushes and removes the oldest key. Caller must hold mu.
func (c *Coalescer) evictOldest() error {
	if len(c.order) == 0 {
		return nil
	}
	key := c.order[0]
	ev := c.window[key]
	c.remove(key)
	if ev != nil {
		return c.flush(ev)
	}
	return nil
}

// remove deletes key from all internal structures. Caller must hold mu.
func (c *Coalescer) remove(key string) {
	delete(c.window, key)
	if t, ok := c.timers[key]; ok {
		t.Stop()
		delete(c.timers, key)
	}
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
}
