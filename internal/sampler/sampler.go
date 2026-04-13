// Package sampler provides probabilistic event sampling for pgstream pipelines.
// It allows a configurable fraction of WAL events to pass through, reducing
// downstream load during high-throughput periods.
package sampler

import (
	"errors"
	"math/rand"
	"sync"

	"github.com/pgstream/pgstream/internal/wal"
)

// Config holds the configuration for the Sampler.
type Config struct {
	// Rate is the fraction of events to allow through, in the range (0.0, 1.0].
	// A rate of 1.0 passes all events; 0.5 passes roughly half.
	Rate float64
}

// Sampler decides whether an event should be forwarded based on a sampling rate.
type Sampler struct {
	cfg  Config
	mu   sync.Mutex
	rng  *rand.Rand
}

// New creates a new Sampler with the given config.
// Returns an error if Rate is not in (0.0, 1.0].
func New(cfg Config) (*Sampler, error) {
	if cfg.Rate <= 0.0 || cfg.Rate > 1.0 {
		return nil, errors.New("sampler: rate must be in the range (0.0, 1.0]")
	}
	return &Sampler{
		cfg: cfg,
		rng: rand.New(rand.NewSource(rand.Int63())),
	}, nil
}

// Sample returns true if the event should be forwarded, false if it should be
// dropped. A nil event is always dropped.
func (s *Sampler) Sample(event *wal.Event) bool {
	if event == nil {
		return false
	}
	if s.cfg.Rate >= 1.0 {
		return true
	}
	s.mu.Lock()
	v := s.rng.Float64()
	s.mu.Unlock()
	return v < s.cfg.Rate
}

// Rate returns the configured sampling rate.
func (s *Sampler) Rate() float64 {
	return s.cfg.Rate
}
