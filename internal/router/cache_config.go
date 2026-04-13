package router

import (
	"errors"
	"time"
)

// DefaultCacheMaxAge is the default max-age when none is specified.
const DefaultCacheMaxAge = 60 * time.Second

// Validate checks the CacheConfig for invalid combinations.
// A nil config is considered valid (the middleware becomes a no-op).
func (c *CacheConfig) Validate() error {
	if c == nil {
		return nil
	}
	if c.NoStore && (c.MaxAge > 0 || c.Private) {
		return errors.New("cache: NoStore is mutually exclusive with MaxAge and Private")
	}
	if c.MaxAge < 0 {
		return errors.New("cache: MaxAge must not be negative")
	}
	return nil
}

// ApplyDefaults fills in zero values with sensible defaults.
func (c *CacheConfig) ApplyDefaults() {
	if c == nil {
		return
	}
	if !c.NoStore && c.MaxAge == 0 {
		c.MaxAge = DefaultCacheMaxAge
	}
}
