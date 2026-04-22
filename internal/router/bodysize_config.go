package router

import "fmt"

const defaultMaxBodyBytes int64 = 1 << 20 // 1 MiB

// Validate checks that the BodySizeConfig is valid.
// A nil config is considered valid (limit disabled).
func (c *BodySizeConfig) Validate() error {
	if c == nil {
		return nil
	}
	if c.MaxBytes < 0 {
		return fmt.Errorf("bodysize: MaxBytes must be >= 0, got %d", c.MaxBytes)
	}
	return nil
}

// DefaultBodySizeConfig returns a BodySizeConfig with sensible defaults.
func DefaultBodySizeConfig() *BodySizeConfig {
	return &BodySizeConfig{MaxBytes: defaultMaxBodyBytes}
}
