// Package healthcheck provides a simple HTTP health check endpoint
// for monitoring the liveness of the pgstream process.
package healthcheck

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Status represents the health status of the service.
type Status struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version,omitempty"`
}

// Server is a lightweight HTTP server exposing a /healthz endpoint.
type Server struct {
	server  *http.Server
	version string
}

// New creates a new health check Server listening on the given address.
func New(addr, version string) *Server {
	mux := http.NewServeMux()
	s := &Server{version: version}
	mux.HandleFunc("/healthz", s.handleHealth)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	return s
}

// Start begins serving health check requests. It blocks until the context is
// cancelled, then performs a graceful shutdown.
func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(shutCtx)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(Status{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
		Version:   s.version,
	}); err != nil {
		// Note: WriteHeader has already been called, so we cannot change the
		// status code at this point. Log the encoding failure instead.
		http.Error(w, "failed to encode health status", http.StatusInternalServerError)
	}
}

// Addr returns the configured listening address of the health check server.
func (s *Server) Addr() string {
	return s.server.Addr
}
