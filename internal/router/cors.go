package router

import (
	"net/http"
	"strconv"
	"strings"
)

// corsHandler wraps an http.Handler with CORS support.
type corsHandler struct {
	next http.Handler
	cfg  *CORSConfig
}

// WithCORS returns middleware that adds CORS headers based on cfg.
// If cfg is nil or has no allowed origins the middleware is a no-op.
func WithCORS(cfg *CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if cfg == nil || len(cfg.AllowedOrigins) == 0 {
			return next
		}
		return &corsHandler{next: next, cfg: cfg}
	}
}

func (h *corsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin != "" && h.isAllowed(origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		if h.cfg.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		if len(h.cfg.AllowedHeaders) > 0 {
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(h.cfg.AllowedHeaders, ", "))
		}
		if len(h.cfg.AllowedMethods) > 0 {
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(h.cfg.AllowedMethods, ", "))
		}
		if h.cfg.MaxAge > 0 {
			w.Header().Set("Access-Control-Max-Age", strconv.Itoa(h.cfg.MaxAge))
		}
	}

	// Handle pre-flight requests.
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	h.next.ServeHTTP(w, r)
}

func (h *corsHandler) isAllowed(origin string) bool {
	for _, o := range h.cfg.AllowedOrigins {
		if o == "*" || strings.EqualFold(o, origin) {
			return true
		}
	}
	return false
}
