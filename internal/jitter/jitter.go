// Package jitter provides utilities for adding randomised jitter to durations,
// useful for spreading retry or polling intervals to avoid thundering-herd problems.
package jitter

import (
	"math/rand"
	"time"
)

// Config holds the configuration for jitter calculation.
type Config struct {
	// Factor is the maximum fraction of the base duration to add as jitter.
	// Must be in the range [0, 1]. Defaults to 0.2 (20%).
	Factor float64
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Factor: 0.2,
	}
}

// Jitter holds the resolved configuration.
type Jitter struct {
	cfg Config
	rng *rand.Rand
}

// New creates a new Jitter instance. If cfg.Factor is zero or negative the
// default factor is used.
func New(cfg Config) (*Jitter, error) {
	if cfg.Factor <= 0 || cfg.Factor > 1 {
		cfg.Factor = DefaultConfig().Factor
	}
	return &Jitter{
		cfg: cfg,
		//nolint:gosec // non-cryptographic randomness is intentional here
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

// Apply returns base plus a random jitter in the range [0, base*Factor].
func (j *Jitter) Apply(base time.Duration) time.Duration {
	if base <= 0 {
		return base
	}
	max := float64(base) * j.cfg.Factor
	delta := time.Duration(j.rng.Float64() * max)
	return base + delta
}

// ApplyFull returns a duration uniformly distributed in [base*(1-Factor), base*(1+Factor)].
func (j *Jitter) ApplyFull(base time.Duration) time.Duration {
	if base <= 0 {
		return base
	}
	spread := float64(base) * j.cfg.Factor
	delta := time.Duration((j.rng.Float64()*2-1)*spread)
	return base + delta
}
