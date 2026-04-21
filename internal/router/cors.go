package router

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// isOriginAllowed reports whether origin is in the allowed list.
func isOriginAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		if a == "*" || a == origin {
			return true
		}
	}
	return false
}

// WithCORS returns middleware that adds CORS headers to responses.
// If cfg is nil or has no origins configured the middleware is a no-op.
func WithCORS(cfg *CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if cfg == nil || len(cfg.AllowedOrigins) == 0 {
			return next
		}
		methods := strings.Join(cfg.AllowedMethods, ", ")
		if methods == "" {
			methods = "GET, POST, PUT, DELETE, OPTIONS"
		}
		headers := strings.Join(cfg.AllowedHeaders, ", ")

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && isOriginAllowed(origin, cfg.AllowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", methods)
				if headers != "" {
					w.Header().Set("Access-Control-Allow-Headers", headers)
				}
				if cfg.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
				if cfg.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%s", strconv.Itoa(cfg.MaxAge)))
				}
			}
			// Handle preflight.
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
