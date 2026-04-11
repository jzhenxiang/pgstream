// Package telemetry provides structured runtime telemetry for pgstream,
// aggregating key operational counters and gauges into a single snapshot
// suitable for logging, metrics export, or health endpoints.
package telemetry

import (
	"sync"
	"sync/atomic"
	"time"
)

// Snapshot holds a point-in-time view of runtime telemetry.
type Snapshot struct {
	EventsReceived  int64     `json:"events_received"`
	EventsProcessed int64     `json:"events_processed"`
	EventsFailed    int64     `json:"events_failed"`
	EventsFiltered  int64     `json:"events_filtered"`
	BytesProcessed  int64     `json:"bytes_processed"`
	Uptime          string    `json:"uptime"`
	CapturedAt      time.Time `json:"captured_at"`
}

// Telemetry tracks operational counters for a running pipeline.
type Telemetry struct {
	mu              sync.Mutex
	start           time.Time
	eventsReceived  atomic.Int64
	eventsProcessed atomic.Int64
	eventsFailed    atomic.Int64
	eventsFiltered  atomic.Int64
	bytesProcessed  atomic.Int64
}

// New creates a new Telemetry instance with the start time set to now.
func New() *Telemetry {
	return &Telemetry{start: time.Now()}
}

// IncReceived increments the received event counter.
func (t *Telemetry) IncReceived() { t.eventsReceived.Add(1) }

// IncProcessed increments the processed event counter.
func (t *Telemetry) IncProcessed() { t.eventsProcessed.Add(1) }

// IncFailed increments the failed event counter.
func (t *Telemetry) IncFailed() { t.eventsFailed.Add(1) }

// IncFiltered increments the filtered event counter.
func (t *Telemetry) IncFiltered() { t.eventsFiltered.Add(1) }

// AddBytes adds n to the bytes-processed counter.
func (t *Telemetry) AddBytes(n int64) { t.bytesProcessed.Add(n) }

// Snapshot returns a consistent point-in-time view of all counters.
func (t *Telemetry) Snapshot() Snapshot {
	t.mu.Lock()
	now := time.Now()
	uptime := now.Sub(t.start).Round(time.Second).String()
	t.mu.Unlock()

	return Snapshot{
		EventsReceived:  t.eventsReceived.Load(),
		EventsProcessed: t.eventsProcessed.Load(),
		EventsFailed:    t.eventsFailed.Load(),
		EventsFiltered:  t.eventsFiltered.Load(),
		BytesProcessed:  t.bytesProcessed.Load(),
		Uptime:          uptime,
		CapturedAt:      now,
	}
}
