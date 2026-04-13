package router

import (
	"net/http"
	"strconv"
	"time"
)

// CacheConfig holds configuration for HTTP response caching headers.
type CacheConfig struct {
	// MaxAge is the max-age directive value in seconds.
	MaxAge time.Duration
	// Private marks the response as private (not storable by shared caches).
	Private bool
	// NoStore disables caching entirely.
	NoStore bool
}

// WithCache is middleware that sets Cache-Control response headers based on cfg.
// If cfg is nil the middleware is a no-op.
func WithCache(cfg *CacheConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if cfg == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var directive string
			switch {
			case cfg.NoStore:
				directive = "no-store"
			case cfg.Private:
				directive = "private, max-age=" + strconv.Itoa(int(cfg.MaxAge.Seconds()))
			default:
				directive = "public, max-age=" + strconv.Itoa(int(cfg.MaxAge.Seconds()))
			}
			w.Header().Set("Cache-Control", directive)
			next.ServeHTTP(w, r)
		})
	}
}
