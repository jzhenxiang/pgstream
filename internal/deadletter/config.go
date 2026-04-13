package deadletter

import "errors"

const DefaultCapacity = 1000

// Config holds configuration for the dead-letter queue.
type Config struct {
	// Capacity is the maximum number of entries held in memory before the
	// oldest entries are evicted. Defaults to DefaultCapacity.
	Capacity int `yaml:"capacity" env:"DLQ_CAPACITY"`
}

// Validate returns an error if the config contains invalid values.
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("deadletter: config must not be nil")
	}
	if c.Capacity < 0 {
		return errors.New("deadletter: capacity must not be negative")
	}
	return nil
}
