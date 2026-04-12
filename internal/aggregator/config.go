package aggregator

import (
	"errors"
	"time"
)

const (
	defaultWindowSize    = 100
	defaultFlushInterval = 5 * time.Second
)

var errNilFlushFunc = errors.New("aggregator: flush function must not be nil")

// Config holds tuning parameters for the Aggregator.
type Config struct {
	// WindowSize is the maximum number of events per table before a flush is
	// triggered. Defaults to 100.
	WindowSize int
	// FlushInterval is how often pending events are flushed regardless of size.
	// Defaults to 5 s.
	FlushInterval time.Duration
}

func (c Config) validate() error {
	if c.WindowSize < 0 {
		return errors.New("aggregator: window size must not be negative")
	}
	if c.FlushInterval < 0 {
		return errors.New("aggregator: flush interval must not be negative")
	}
	return nil
}

func (c Config) windowSize() int {
	if c.WindowSize == 0 {
		return defaultWindowSize
	}
	return c.WindowSize
}

func (c Config) flushInterval() time.Duration {
	if c.FlushInterval == 0 {
		return defaultFlushInterval
	}
	return c.FlushInterval
}
