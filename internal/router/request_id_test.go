package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithRequestID_GeneratesIDWhenAbsent(t *testing.T) {
	var capturedID string
	handler := WithRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if capturedID == "" {
		t.Fatal("expected a non-empty request ID in context")
	}
	if rec.Header().Get(RequestIDHeader) != capturedID {
		t.Errorf("response header mismatch: got %q, want %q",
			rec.Header().Get(RequestIDHeader), capturedID)
	}
}

func TestWithRequestID_ReusesClientSuppliedID(t *testing.T) {
	const clientID = "my-trace-id-123"
	var capturedID string
	handler := WithRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, clientID)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if capturedID != clientID {
		t.Errorf("expected context ID %q, got %q", clientID, capturedID)
	}
	if rec.Header().Get(RequestIDHeader) != clientID {
		t.Errorf("expected response header %q, got %q", clientID, rec.Header().Get(RequestIDHeader))
	}
}

func TestWithRequestID_UniquePerRequest(t *testing.T) {
	ids := make([]string, 5)
	handler := WithRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := range ids {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		ids[i] = rec.Header().Get(RequestIDHeader)
	}

	seen := make(map[string]struct{})
	for _, id := range ids {
		if _, dup := seen[id]; dup {
			t.Errorf("duplicate request ID generated: %q", id)
		}
		seen[id] = struct{}{}
	}
}

func TestRequestIDFromContext_MissingKey_ReturnsEmpty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if id := RequestIDFromContext(req.Context()); id != "" {
		t.Errorf("expected empty string, got %q", id)
	}
}
