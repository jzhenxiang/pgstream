package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
)

func TestWithVersion_DefaultInfo(t *testing.T) {
	h := WithVersion(nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/version", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var info BuildInfo
	if err := json.NewDecoder(rec.Body).Decode(&info); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if info.Version != "dev" {
		t.Errorf("expected version=dev, got %q", info.Version)
	}
	if info.GoVersion != runtime.Version() {
		t.Errorf("expected go version %q, got %q", runtime.Version(), info.GoVersion)
	}
}

func TestWithVersion_CustomInfo(t *testing.T) {
	custom := &BuildInfo{
		Version:   "1.2.3",
		Commit:    "abc123",
		BuildTime: "2024-01-01T00:00:00Z",
	}
	h := WithVersion(custom)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/version", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var info BuildInfo
	if err := json.NewDecoder(rec.Body).Decode(&info); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if info.Version != "1.2.3" {
		t.Errorf("expected version=1.2.3, got %q", info.Version)
	}
	if info.Commit != "abc123" {
		t.Errorf("expected commit=abc123, got %q", info.Commit)
	}
	if info.GoVersion != runtime.Version() {
		t.Errorf("expected go version to be populated, got %q", info.GoVersion)
	}
}

func TestWithVersion_ContentTypeJSON(t *testing.T) {
	h := WithVersion(nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/version", nil))

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}
