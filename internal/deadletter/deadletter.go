package deadletter

import (
	"context"
	"sync"
	"time"

	"github.com/pgstream/pgstream/internal/wal"
)

// Entry holds a failed WAL event along with metadata about why it failed.
type Entry struct {
	Event     *wal.Event
	Err       error
	Attempts  int
	FailedAt  time.Time
}

// Queue is a thread-safe in-memory dead-letter queue with optional overflow
// eviction when the capacity is exceeded.
type Queue struct {
	mu       sync.Mutex
	entries  []*Entry
	capacity int
}

// New creates a new Queue with the given capacity. If capacity is zero or
// negative the DefaultCapacity is used.
func New(capacity int) *Queue {
	if capacity <= 0 {
		capacity = DefaultCapacity
	}
	return &Queue{capacity: capacity}
}

// Push appends a failed event to the queue. If the queue is at capacity the
// oldest entry is evicted to make room (FIFO eviction).
func (q *Queue) Push(ctx context.Context, event *wal.Event, err error, attempts int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	entry := &Entry{
		Event:    event,
		Err:      err,
		Attempts: attempts,
		FailedAt: time.Now().UTC(),
	}

	if len(q.entries) >= q.capacity {
		q.entries = q.entries[1:]
	}
	q.entries = append(q.entries, entry)
}

// Drain returns all current entries and clears the queue.
func (q *Queue) Drain() []*Entry {
	q.mu.Lock()
	defer q.mu.Unlock()

	out := make([]*Entry, len(q.entries))
	copy(out, q.entries)
	q.entries = q.entries[:0]
	return out
}

// Len returns the current number of entries.
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.entries)
}

// Peek returns a snapshot of the entries without clearing the queue.
func (q *Queue) Peek() []*Entry {
	q.mu.Lock()
	defer q.mu.Unlock()

	out := make([]*Entry, len(q.entries))
	copy(out, q.entries)
	return out
}
