// Package cursor tracks the current read position within a WAL stream,
// providing a lightweight wrapper around an LSN value with thread-safe
// advance semantics.
package cursor

import (
	"fmt"
	"sync"
)

// LSN represents a PostgreSQL Log Sequence Number.
type LSN uint64

// String returns the standard X/Y hex representation of an LSN.
func (l LSN) String() string {
	return fmt.Sprintf("%X/%X", uint32(l>>32), uint32(l))
}

// Cursor holds the current and high-water-mark LSN positions.
type Cursor struct {
	mu      sync.RWMutex
	current LSN
	hwm     LSN
}

// New returns a Cursor initialised to the given starting LSN.
func New(start LSN) *Cursor {
	return &Cursor{current: start, hwm: start}
}

// Current returns the most recently committed LSN.
func (c *Cursor) Current() LSN {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.current
}

// HighWaterMark returns the highest LSN ever observed.
func (c *Cursor) HighWaterMark() LSN {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hwm
}

// Advance moves the cursor forward to lsn. It is a no-op when lsn is less
// than or equal to the current position. Returns true if the position changed.
func (c *Cursor) Advance(lsn LSN) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if lsn <= c.current {
		return false
	}
	c.current = lsn
	if lsn > c.hwm {
		c.hwm = lsn
	}
	return true
}

// Reset sets the cursor back to the given LSN without updating the HWM.
func (c *Cursor) Reset(lsn LSN) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.current = lsn
}
