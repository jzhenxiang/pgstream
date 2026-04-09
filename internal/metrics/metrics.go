package metrics

import (
	"sync/atomic"
	"time"
)

// Metrics holds runtime counters for the pgstream pipeline.
type Metrics struct {
	MessagesReceived  atomic.Int64
	MessagesProcessed atomic.Int64
	MessagesFailed    atomic.Int64
	BytesProcessed    atomic.Int64
	StartTime         time.Time
}

// New creates a new Metrics instance with the start time set to now.
func New() *Metrics {
	return &Metrics{
		StartTime: time.Now(),
	}
}

// RecordReceived increments the received message counter.
func (m *Metrics) RecordReceived() {
	m.MessagesReceived.Add(1)
}

// RecordProcessed increments the processed message counter and adds byte count.
func (m *Metrics) RecordProcessed(bytes int64) {
	m.MessagesProcessed.Add(1)
	m.BytesProcessed.Add(bytes)
}

// RecordFailed increments the failed message counter.
func (m *Metrics) RecordFailed() {
	m.MessagesFailed.Add(1)
}

// Snapshot returns a point-in-time copy of the current metric values.
func (m *Metrics) Snapshot() Snapshot {
	return Snapshot{
		MessagesReceived:  m.MessagesReceived.Load(),
		MessagesProcessed: m.MessagesProcessed.Load(),
		MessagesFailed:    m.MessagesFailed.Load(),
		BytesProcessed:    m.BytesProcessed.Load(),
		UptimeSeconds:     int64(time.Since(m.StartTime).Seconds()),
	}
}

// Snapshot is an immutable point-in-time view of Metrics.
type Snapshot struct {
	MessagesReceived  int64
	MessagesProcessed int64
	MessagesFailed    int64
	BytesProcessed    int64
	UptimeSeconds     int64
}
