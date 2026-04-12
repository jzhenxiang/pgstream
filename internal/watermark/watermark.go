// Package watermark tracks the high-water mark of processed WAL LSN positions,
// enabling safe advancement of the replication slot's confirmed flush position.
package watermark

import (
	"fmt"
	"sync"
)

// LSN represents a PostgreSQL Log Sequence Number.
type LSN uint64

// String returns the standard PostgreSQL LSN format (X/YYYYYYYY).
func (l LSN) String() string {
	return fmt.Sprintf("%X/%08X", uint32(l>>32), uint32(l))
}

// Watermark tracks the highest contiguously confirmed LSN.
// It maintains an in-flight set of pending LSNs and advances the
// confirmed mark only when there are no gaps below the minimum pending.
type Watermark struct {
	mu        sync.Mutex
	confirmed LSN
	pending   map[LSN]struct{}
}

// New creates a new Watermark starting at the given initial LSN.
func New(initial LSN) *Watermark {
	return &Watermark{
		confirmed: initial,
		pending:   make(map[LSN]struct{}),
	}
}

// Track registers an LSN as in-flight (not yet confirmed).
func (w *Watermark) Track(lsn LSN) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.pending[lsn] = struct{}{}
}

// Confirm marks the given LSN as successfully processed and advances
// the confirmed watermark if possible.
func (w *Watermark) Confirm(lsn LSN) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.pending, lsn)
	if lsn > w.confirmed && len(w.pending) == 0 {
		w.confirmed = lsn
		return
	}
	// Advance to just below the lowest still-pending LSN.
	if lsn > w.confirmed {
		lowest := lsn
		for p := range w.pending {
			if p < lowest {
				lowest = p
			}
		}
		if lowest > w.confirmed+1 {
			w.confirmed = lowest - 1
		}
	}
}

// Confirmed returns the current high-water mark LSN.
func (w *Watermark) Confirmed() LSN {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.confirmed
}

// PendingCount returns the number of in-flight LSNs not yet confirmed.
func (w *Watermark) PendingCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.pending)
}
