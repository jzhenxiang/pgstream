package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithTracing_NilSink_IsNoOp(t *testing.T) {
	mw := WithTracing(nil)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithTracing_RecordsEntry(t *testing.T) {
	var got TraceEntry
	mw := WithTracing(func(e TraceEntry) { got = e })
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/events", nil))
	if got.Method != http.MethodPost {
		t.Errorf("expected POST, got %s", got.Method)
	}
	if got.Path != "/events" {
		t.Errorf("expected /events, got %s", got.Path)
	}
	if got.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", got.StatusCode)
	}
	if got.Duration <= 0 {
		t.Error("expected positive duration")
	}
	if got.RequestID == "" {
		t.Error("expected non-empty request ID")
	}
}

func TestWithTracing_DefaultStatusCode(t *testing.T) {
	var got TraceEntry
	mw := WithTracing(func(e TraceEntry) { got = e })
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// WriteHeader never called explicitly
		_, _ = w.Write([]byte("ok"))
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if got.StatusCode != http.StatusOK {
		t.Errorf("expected default 200, got %d", got.StatusCode)
	}
}
