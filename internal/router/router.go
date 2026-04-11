// Package router wires together the HTTP endpoints exposed by pgstream,
// including the health-check and any webhook-receiver routes.
package router

import (
	"log/slog"
	"net/http"

	"github.com/pgstream/pgstream/internal/healthcheck"
	"github.com/pgstream/pgstream/internal/middleware"
)

// Config holds router-level configuration.
type Config struct {
	// SigningSecret, when non-empty, enables HMAC signature verification on
	// inbound webhook requests.
	SigningSecret string
	// SignatureHeader is the HTTP header that carries the HMAC signature.
	// Defaults to "X-Signature" when empty.
	SignatureHeader string
	Logger          *slog.Logger
}

// New builds and returns an http.Handler that composes all pgstream HTTP
// endpoints. hc must not be nil.
func New(cfg Config, hc *healthcheck.HealthCheck) (http.Handler, error) {
	mux := http.NewServeMux()

	// Health-check endpoint.
	mux.HandleFunc("/healthz", hc.HandleHealth)

	var handler http.Handler = mux

	// Wrap with structured logger middleware.
	handler = middleware.Logger(cfg.Logger, handler)

	// Wrap with panic recovery.
	handler = middleware.Recovery(cfg.Logger, handler)

	// Optionally wrap with HMAC signature verification.
	if cfg.SigningSecret != "" {
		signer, err := middleware.NewSigner(cfg.SigningSecret, cfg.SignatureHeader)
		if err != nil {
			return nil, err
		}
		handler = signer.Verify(handler)
	}

	return handler, nil
}
