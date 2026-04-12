// Package window provides a sliding time-window counter used to track
// event rates over a rolling duration (e.g. events per minute).
package window

import (
	"sync"
	"time"
)

// entry holds a single timestamped count.
type entry struct {
	at    time.Time
	count int64
}

// Window is a thread-safe sliding time-window counter.
type Window struct {
	mu       sync.Mutex
	buckets  []entry
	size     time.Duration
	granule  time.Duration
}

// New creates a Window with the given total size and bucket granularity.
// granule controls how finely events are bucketed; size must be > granule.
func New(size, granule time.Duration) (*Window, error) {
	if size <= 0 {
		return nil, ErrInvalidSize
	}
	if granule <= 0 || granule > size {
		return nil, ErrInvalidGranule
	}
	return &Window{
		size:    size,
		granule: granule,
	}, nil
}

// Add records n events at the current time.
func (w *Window) Add(n int64) {
	now := time.Now().Truncate(w.granule)
	w.mu.Lock()
	defer w.mu.Unlock()
	w.evict(now)
	if len(w.buckets) > 0 && w.buckets[len(w.buckets)-1].at.Equal(now) {
		w.buckets[len(w.buckets)-1].count += n
		return
	}
	w.buckets = append(w.buckets, entry{at: now, count: n})
}

// Count returns the total number of events recorded within the window.
func (w *Window) Count() int64 {
	now := time.Now().Truncate(w.granule)
	w.mu.Lock()
	defer w.mu.Unlock()
	w.evict(now)
	var total int64
	for _, b := range w.buckets {
		total += b.count
	}
	return total
}

// Reset clears all recorded events.
func (w *Window) Reset() {
	w.mu.Lock()
	w.buckets = w.buckets[:0]
	w.mu.Unlock()
}

// evict removes buckets older than the window size. Must be called with mu held.
func (w *Window) evict(now time.Time) {
	cutoff := now.Add(-w.size)
	i := 0
	for i < len(w.buckets) && !w.buckets[i].at.After(cutoff) {
		i++
	}
	w.buckets = w.buckets[i:]
}
