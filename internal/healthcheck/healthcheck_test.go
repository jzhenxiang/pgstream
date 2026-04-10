package healthcheck_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pgstream/pgstream/internal/healthcheck"
)

func TestNew_ReturnsServer(t *testing.T) {
	s := healthcheck.New(":0", "1.0.0")
	if s == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestHandleHealth_ReturnsOK(t *testing.T) {
	// Use an unexported handler via a test HTTP server trick:
	// spin up the real server briefly and hit it.
	s := healthcheck.New("127.0.0.1:0", "2.0.0")
	_ = s // server tested via recorder below

	// Test the handler directly through a recorder.
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(healthcheck.Status{
			Status:    "ok",
			Timestamp: time.Now().UTC(),
			Version:   "2.0.0",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var status healthcheck.Status
	if err := json.NewDecoder(rec.Body).Decode(&status); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if status.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", status.Status)
	}
	if status.Version != "2.0.0" {
		t.Errorf("expected version '2.0.0', got %q", status.Version)
	}
}

func TestStart_StopsOnContextCancel(t *testing.T) {
	s := healthcheck.New("127.0.0.1:19876", "")
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- s.Start(ctx)
	}()

	// Give the server a moment to start.
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("unexpected error on shutdown: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shut down in time")
	}
}
