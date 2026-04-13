package router

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSConfig holds configuration for the CORS middleware.
type CORSConfig struct {
	// AllowedOrigins is a list of origins that are allowed. Use ["*"] to allow all.
	AllowedOrigins []string
	// AllowedMethods is a list of HTTP methods allowed. Defaults to GET, POST, OPTIONS.
	AllowedMethods []string
	// AllowedHeaders is a list of HTTP headers allowed.
	AllowedHeaders []string
	// MaxAge is the value (in seconds) for the Access-Control-Max-Age header.
	MaxAge int
}

func (c *CORSConfig) isOriginAllowed(origin string) bool {
	for _, o := range c.AllowedOrigins {
		if o == "*" || strings.EqualFold(o, origin) {
			return true
		}
	}
	return false
}

// WithCORS returns a middleware that applies CORS headers based on cfg.
// If cfg is nil or AllowedOrigins is empty the middleware is a no-op.
func WithCORS(cfg *CORSConfig) func(http.Handler) http.Handler {
	if cfg == nil || len(cfg.AllowedOrigins) == 0 {
		return func(next http.Handler) http.Handler { return next }
	}

	methods := strings.Join(cfg.AllowedMethods, ", ")
	if methods == "" {
		methods = "GET, POST, OPTIONS"
	}
	headers := strings.Join(cfg.AllowedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && cfg.isOriginAllowed(origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", methods)
				if headers != "" {
					w.Header().Set("Access-Control-Allow-Headers", headers)
				}
				if cfg.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAge))
				}
				// Vary header ensures caches correctly store per-origin responses.
				w.Header().Add("Vary", "Origin")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
