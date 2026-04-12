// Package enricher attaches additional metadata fields to WAL events
// before they are forwarded to a sink.
package enricher

import (
	"fmt"
	"os"
	"time"

	"github.com/pgstream/pgstream/internal/wal"
)

// Config holds the enrichment rules.
type Config struct {
	// StaticFields are key/value pairs added to every event.
	StaticFields map[string]string
	// AddHostname appends a "_hostname" field when true.
	AddHostname bool
	// AddTimestamp appends an "_enriched_at" field when true.
	AddTimestamp bool
}

// Enricher adds metadata to WAL events.
type Enricher struct {
	cfg      Config
	hostname string
	now      func() time.Time
}

// New returns an Enricher configured with cfg.
func New(cfg Config) (*Enricher, error) {
	var host string
	if cfg.AddHostname {
		h, err := os.Hostname()
		if err != nil {
			return nil, fmt.Errorf("enricher: resolve hostname: %w", err)
		}
		host = h
	}
	return &Enricher{
		cfg:      cfg,
		hostname: host,
		now:      time.Now,
	}, nil
}

// Apply returns a shallow copy of event with the configured metadata fields
// merged into its Metadata map. If event is nil, nil is returned.
func (e *Enricher) Apply(event *wal.Event) *wal.Event {
	if event == nil {
		return nil
	}

	out := *event
	out.Metadata = cloneMetadata(event.Metadata)

	for k, v := range e.cfg.StaticFields {
		out.Metadata[k] = v
	}
	if e.cfg.AddHostname && e.hostname != "" {
		out.Metadata["_hostname"] = e.hostname
	}
	if e.cfg.AddTimestamp {
		out.Metadata["_enriched_at"] = e.now().UTC().Format(time.RFC3339Nano)
	}
	return &out
}

func cloneMetadata(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src)+4)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
