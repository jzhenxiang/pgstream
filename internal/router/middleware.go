package router

import (
	"net/http"
	"time"

	"github.com/pgstream/pgstream/internal/middleware"
)

// Chain wraps a handler with a sequence of middleware functions applied
// from outermost to innermost.
func Chain(h http.Handler, mw ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}

// WithTimeout wraps a handler so that requests exceeding d are cancelled.
func WithTimeout(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if d > 0 {
				var cancel func()
				ctx, cancel = withDeadline(ctx, d)
				defer cancel()
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// WithSigning returns a middleware that verifies HMAC signatures when a
// non-empty secret is provided. If the secret is empty the middleware is a
// no-op pass-through.
func WithSigning(secret string) (func(http.Handler) http.Handler, error) {
	if secret == "" {
		return func(next http.Handler) http.Handler { return next }, nil
	}
	signer, err := middleware.NewSigner(secret, "")
	if err != nil {
		return nil, err
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := signer.Verify(r); err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}, nil
}
