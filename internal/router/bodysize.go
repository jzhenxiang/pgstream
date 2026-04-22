package router

import (
	"net/http"
)

// BodySizeConfig controls the maximum allowed request body size.
type BodySizeConfig struct {
	// MaxBytes is the maximum number of bytes allowed in a request body.
	// Zero or negative values disable the limit.
	MaxBytes int64
}

// WithBodySize returns middleware that rejects requests whose body exceeds
// MaxBytes. When the limit is not set the middleware is a no-op.
func WithBodySize(cfg *BodySizeConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if cfg == nil || cfg.MaxBytes <= 0 {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
