package router

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithHealthz_DefaultEndpoint_Returns200(t *testing.T) {
	mux := http.NewServeMux()
	WithHealthz(mux, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithHealthz_MethodNotAllowed(t *testing.T) {
	mux := http.NewServeMux()
	WithHealthz(mux, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestWithHealthz_CustomEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	WithHealthz(mux, &HealthzConfig{Endpoint: "/ready"})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithHealthz_AllChecksPass_ReturnsOK(t *testing.T) {
	mux := http.NewServeMux()
	WithHealthz(mux, &HealthzConfig{
		Checks: map[string]func() error{
			"db": func() error { return nil },
		},
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp healthzResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Status != "ok" {
		t.Fatalf("expected ok, got %s", resp.Status)
	}
	if resp.Checks["db"] != "ok" {
		t.Fatalf("expected db ok, got %s", resp.Checks["db"])
	}
}

func TestWithHealthz_FailingCheck_ReturnsDegraded(t *testing.T) {
	mux := http.NewServeMux()
	WithHealthz(mux, &HealthzConfig{
		Checks: map[string]func() error{
			"kafka": func() error { return errors.New("connection refused") },
		},
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}

	var resp healthzResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Status != "degraded" {
		t.Fatalf("expected degraded, got %s", resp.Status)
	}
}
