package router

import (
	"net/http"
	"strings"
)

// RedirectConfig controls how the redirect middleware behaves.
type RedirectConfig struct {
	// HTTPSOnly redirects all plain-HTTP requests to their HTTPS equivalent.
	HTTPSOnly bool

	// TrailingSlash removes a trailing slash from the request path before
	// passing it to the next handler (e.g. /foo/ → /foo).
	// If the path is exactly "/" it is left unchanged.
	TrailingSlash bool

	// StatusCode is the HTTP redirect status to use. Defaults to 301.
	StatusCode int
}

func (c *RedirectConfig) statusCode() int {
	if c == nil || c.StatusCode == 0 {
		return http.StatusMovedPermanently
	}
	return c.StatusCode
}

// WithRedirect returns middleware that enforces URL normalisation rules
// described by cfg. A nil config is a no-op.
//
// Rules are applied in the following order:
//  1. Trailing-slash removal (permanent redirect).
//  2. HTTP → HTTPS upgrade (permanent redirect).
//
// Both rules may be active simultaneously; the first matching rule wins.
func WithRedirect(cfg *RedirectConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if cfg == nil {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Strip trailing slash.
			if cfg.TrailingSlash {
				path := r.URL.Path
				if path != "/" && strings.HasSuffix(path, "/") {
					r.URL.Path = strings.TrimRight(path, "/")
					http.Redirect(w, r, r.URL.String(), cfg.statusCode())
					return
				}
			}

			// 2. Force HTTPS.
			if cfg.HTTPSOnly && r.TLS == nil {
				// Honour the X-Forwarded-Proto header set by a TLS-terminating
				// reverse proxy so that the middleware works correctly behind
				// load-balancers.
				proto := r.Header.Get("X-Forwarded-Proto")
				if !strings.EqualFold(proto, "https") {
					target := "https://" + r.Host + r.URL.RequestURI()
					http.Redirect(w, r, target, cfg.statusCode())
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
