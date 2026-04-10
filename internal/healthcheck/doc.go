// Package healthcheck implements a minimal HTTP health check server for
// pgstream. It exposes a single GET /healthz endpoint that returns a JSON
// payload with the current status, timestamp, and optional version string.
//
// Usage:
//
//	s := healthcheck.New(":8080", "v1.2.3")
//	if err := s.Start(ctx); err != nil {
//		log.Fatal(err)
//	}
//
// The server shuts down gracefully when the provided context is cancelled.
package healthcheck
