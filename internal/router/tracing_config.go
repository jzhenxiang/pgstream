package router

import "errors"

// TracingConfig controls the behaviour of the tracing middleware.
type TracingConfig struct {
	// Enabled turns tracing on or off.
	Enabled bool

	// MaxEntries caps the number of entries held in an in-memory store.
	// Zero means no in-memory store is used.
	MaxEntries int
}

// Validate returns an error if the config is invalid.
func (c *TracingConfig) Validate() error {
	if c == nil {
		return errors.New("tracing: config must not be nil")
	}
	if c.MaxEntries < 0 {
		return errors.New("tracing: MaxEntries must be >= 0")
	}
	return nil
}

// DefaultTracingConfig returns a TracingConfig with sensible defaults.
func DefaultTracingConfig() *TracingConfig {
	return &TracingConfig{
		Enabled:    true,
		MaxEntries: 1000,
	}
}
