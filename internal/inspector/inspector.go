// Package inspector provides a read-only view of the current pipeline state,
// exposing counters, last-seen LSN, and sink health for diagnostic purposes.
package inspector

import (
	"sync"
	"time"

	"github.com/your-org/pgstream/internal/lsn"
)

// State is a snapshot of the pipeline at a point in time.
type State struct {
	LastLSN     lsn.LSN   `json:"last_lsn"`
	Received    uint64    `json:"received"`
	Processed   uint64    `json:"processed"`
	Failed      uint64    `json:"failed"`
	SinkHealthy bool      `json:"sink_healthy"`
	CapturedAt  time.Time `json:"captured_at"`
}

// Inspector tracks live pipeline counters and exposes a point-in-time snapshot.
type Inspector struct {
	mu          sync.RWMutex
	lastLSN     lsn.LSN
	received    uint64
	processed   uint64
	failed      uint64
	sinkHealthy bool
}

// New returns an initialised Inspector with sink assumed healthy.
func New() *Inspector {
	return &Inspector{sinkHealthy: true}
}

// RecordReceived increments the received counter and updates the last seen LSN.
func (i *Inspector) RecordReceived(l lsn.LSN) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.received++
	if l.After(i.lastLSN) {
		i.lastLSN = l
	}
}

// RecordProcessed increments the processed counter.
func (i *Inspector) RecordProcessed() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.processed++
}

// RecordFailed increments the failed counter and marks the sink as unhealthy.
func (i *Inspector) RecordFailed() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.failed++
	i.sinkHealthy = false
}

// MarkSinkHealthy resets the sink health flag.
func (i *Inspector) MarkSinkHealthy() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.sinkHealthy = true
}

// Snapshot returns a consistent copy of the current state.
func (i *Inspector) Snapshot() State {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return State{
		LastLSN:     i.lastLSN,
		Received:    i.received,
		Processed:   i.processed,
		Failed:      i.failed,
		SinkHealthy: i.sinkHealthy,
		CapturedAt:  time.Now().UTC(),
	}
}
