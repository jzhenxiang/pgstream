package router

import (
	"net/http"
	"time"
)

// TimeoutConfig holds configuration for the request timeout middleware.
type TimeoutConfig struct {
	// Duration is the maximum time allowed for a handler to complete.
	// A zero value disables the timeout.
	Duration time.Duration

	// Message is the response body sent when the timeout is exceeded.
	// Defaults to "request timeout".
	Message string
}

func (c *TimeoutConfig) message() string {
	if c == nil || c.Message == "" {
		return "request timeout"
	}
	return c.Message
}

// WithRequestTimeout wraps a handler with an HTTP-level timeout. If cfg is nil
// or cfg.Duration is zero the handler is returned unchanged.
func WithRequestTimeout(cfg *TimeoutConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if cfg == nil || cfg.Duration == 0 {
			return next
		}
		return http.TimeoutHandler(next, cfg.Duration, cfg.message())
	}
}
