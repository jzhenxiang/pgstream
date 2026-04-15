package router

import (
	"errors"
	"time"
)

// DefaultTimeoutDuration is used when no explicit duration is configured.
const DefaultTimeoutDuration = 30 * time.Second

// Validate checks the TimeoutConfig and applies defaults where appropriate.
// A nil receiver is valid and represents a disabled timeout.
func (c *TimeoutConfig) Validate() error {
	if c == nil {
		return nil
	}
	if c.Duration < 0 {
		return errors.New("timeout: duration must not be negative")
	}
	return nil
}

// DefaultTimeoutConfig returns a TimeoutConfig with sensible defaults.
func DefaultTimeoutConfig() *TimeoutConfig {
	return &TimeoutConfig{
		Duration: DefaultTimeoutDuration,
		Message:  "request timeout",
	}
}
