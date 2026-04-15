package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWithMetrics_NilStore_IsNoOp(t *testing.T) {
	mw := WithMetrics(nil)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithMetrics_RecordsRequest(t *testing.T) {
	m := newRequestMetrics()
	mw := WithMetrics(m)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ping", nil))
	snap := m.snapshot()
	if snap["/ping"]["requests"] != "1" {
		t.Fatalf("expected 1 request, got %s", snap["/ping"]["requests"])
	}
	if snap["/ping"]["errors"] != "0" {
		t.Fatalf("expected 0 errors, got %s", snap["/ping"]["errors"])
	}
}

func TestWithMetrics_RecordsError(t *testing.T) {
	m := newRequestMetrics()
	mw := WithMetrics(m)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/fail", nil))
	snap := m.snapshot()
	if snap["/fail"]["errors"] != "1" {
		t.Fatalf("expected 1 error, got %s", snap["/fail"]["errors"])
	}
}

func TestRecord_AvgLatency(t *testing.T) {
	m := newRequestMetrics()
	m.record("/test", 200, 100*time.Millisecond)
	m.record("/test", 200, 200*time.Millisecond)
	snap := m.snapshot()
	if snap["/test"]["avg_latency_ms"] != "150" {
		t.Fatalf("expected avg 150ms, got %s", snap["/test"]["avg_latency_ms"])
	}
}

func TestMetricsHandler_MethodNotAllowed(t *testing.T) {
	m := newRequestMetrics()
	h := metricsHandler(m)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/metrics", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestMetricsHandler_ReturnsJSON(t *testing.T) {
	m := newRequestMetrics()
	m.record("/foo", 200, 10*time.Millisecond)
	h := metricsHandler(m)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
}
