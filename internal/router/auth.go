package router

import (
	"net/http"
	"strings"
)

// AuthConfig holds configuration for bearer token authentication middleware.
type AuthConfig struct {
	// Tokens is the set of valid bearer tokens. At least one must be provided.
	Tokens []string
	// Realm is the WWW-Authenticate realm value returned on 401 responses.
	Realm string
}

// Validate returns an error if the AuthConfig is invalid.
func (c *AuthConfig) Validate() error {
	if c == nil {
		return nil
	}
	for _, t := range c.Tokens {
		if strings.TrimSpace(t) != "" {
			return nil
		}
	}
	return errAuthNoTokens
}

var errAuthNoTokens = authError("auth: at least one non-blank token must be provided")

type authError string

func (e authError) Error() string { return string(e) }

// WithAuth returns a middleware that enforces bearer token authentication.
// Requests without a valid Authorization: Bearer <token> header receive 401.
// If cfg is nil or contains no tokens the middleware is a no-op.
func WithAuth(cfg *AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if cfg == nil || cfg.Validate() != nil {
			return next
		}
		allowed := make(map[string]struct{}, len(cfg.Tokens))
		for _, t := range cfg.Tokens {
			if s := strings.TrimSpace(t); s != "" {
				allowed[s] = struct{}{}
			}
		}
		realm := cfg.Realm
		if realm == "" {
			realm = "pgstream"
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearer(r)
			if _, ok := allowed[token]; !ok {
				w.Header().Set("WWW-Authenticate", `Bearer realm="`+realm+`"`)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func extractBearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(h, prefix) {
		return ""
	}
	return strings.TrimSpace(h[len(prefix):])
}
