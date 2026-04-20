package router

import "errors"

// AuditConfig controls the in-memory audit log sink.
type AuditConfig struct {
	// MaxEntries is the maximum number of entries kept in memory.
	// Zero means the audit log is disabled.
	MaxEntries int
}

// Validate returns an error if the configuration is invalid.
func (c *AuditConfig) Validate() error {
	if c == nil {
		return errors.New("audit: config must not be nil")
	}
	if c.MaxEntries < 0 {
		return errors.New("audit: MaxEntries must be >= 0")
	}
	return nil
}

// DefaultAuditConfig returns a sensible default configuration.
func DefaultAuditConfig() *AuditConfig {
	return &AuditConfig{MaxEntries: 1000}
}

// inMemoryAuditSink is a bounded, in-memory AuditSink used for testing and
// lightweight deployments.
type inMemoryAuditSink struct {
	max     int
	entries []AuditEntry
}

// NewInMemoryAuditSink creates an AuditSink that keeps up to max entries.
// When the buffer is full the oldest entry is evicted.
func NewInMemoryAuditSink(max int) AuditSink {
	if max <= 0 {
		max = DefaultAuditConfig().MaxEntries
	}
	return &inMemoryAuditSink{max: max}
}

func (s *inMemoryAuditSink) Record(e AuditEntry) {
	if len(s.entries) >= s.max {
		s.entries = s.entries[1:]
	}
	s.entries = append(s.entries, e)
}

// Entries returns a copy of the recorded audit entries.
func (s *inMemoryAuditSink) Entries() []AuditEntry {
	out := make([]AuditEntry, len(s.entries))
	copy(out, s.entries)
	return out
}
