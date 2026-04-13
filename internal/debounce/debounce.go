// Package debounce provides a debouncer that delays processing of events
// until a quiet period has elapsed, collapsing rapid successive updates
// into a single emission.
package debounce

import (
	"context"
	"sync"
	"time"

	"github.com/pgstream/pgstream/internal/wal"
)

// DefaultQuietPeriod is the default duration to wait after the last event
// before flushing.
const DefaultQuietPeriod = 200 * time.Millisecond

// Debouncer collapses rapid successive WAL events into a single call to Send.
type Debouncer struct {
	mu          sync.Mutex
	quietPeriod time.Duration
	timer       *time.Timer
	pending     []*wal.Event
	send        func([]*wal.Event) error
}

// New creates a Debouncer that waits quietPeriod after the last received event
// before invoking send. If quietPeriod is zero, DefaultQuietPeriod is used.
func New(quietPeriod time.Duration, send func([]*wal.Event) error) (*Debouncer, error) {
	if send == nil {
		return nil, errNilSend
	}
	if quietPeriod <= 0 {
		quietPeriod = DefaultQuietPeriod
	}
	return &Debouncer{
		quietPeriod: quietPeriod,
		send:        send,
	}, nil
}

// Add accepts an event and resets the quiet-period timer. If the timer fires
// before another event arrives, all pending events are flushed via send.
// Add is safe for concurrent use.
func (d *Debouncer) Add(ctx context.Context, event *wal.Event) {
	if event == nil {
		return
	}
	d.mu.Lock()
	d.pending = append(d.pending, event)
	if d.timer != nil {
		d.timer.Reset(d.quietPeriod)
	} else {
		d.timer = time.AfterFunc(d.quietPeriod, func() {
			d.flush(ctx)
		})
	}
	d.mu.Unlock()
}

// Flush immediately sends any pending events regardless of the quiet period.
func (d *Debouncer) Flush(ctx context.Context) error {
	return d.flush(ctx)
}

func (d *Debouncer) flush(ctx context.Context) error {
	d.mu.Lock()
	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}
	events := d.pending
	d.pending = nil
	d.mu.Unlock()

	if len(events) == 0 {
		return nil
	}
	return d.send(events)
}
