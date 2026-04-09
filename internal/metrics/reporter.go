package metrics

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

// Reporter periodically logs a metrics snapshot.
type Reporter struct {
	metrics  *Metrics
	interval time.Duration
}

// NewReporter creates a Reporter that logs metrics at the given interval.
func NewReporter(m *Metrics, interval time.Duration) *Reporter {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &Reporter{
		metrics:  m,
		interval: interval,
	}
}

// Run starts the periodic reporting loop. It blocks until ctx is cancelled.
func (r *Reporter) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.report()
			return
		case <-ticker.C:
			r.report()
		}
	}
}

func (r *Reporter) report() {
	snap := r.metrics.Snapshot()
	b, err := json.Marshal(snap)
	if err != nil {
		log.Printf("[pgstream] metrics: failed to marshal snapshot: %v", err)
		return
	}
	log.Printf("[pgstream] metrics: %s", string(b))
}
