package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestWithRateLimit_ZeroWindow_PassesThrough(t *testing.T) {
	mw := WithRateLimit(0, 5)
	h := mw(http.HandlerFunc(okHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithRateLimit_ZeroMaxReqs_PassesThrough(t *testing.T) {
	mw := WithRateLimit(time.Second, 0)
	h := mw(http.HandlerFunc(okHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithRateLimit_AllowsUpToMax(t *testing.T) {
	mw := WithRateLimit(time.Second, 3)
	h := mw(http.HandlerFunc(okHandler))

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}
}

func TestWithRateLimit_BlocksOverMax(t *testing.T) {
	mw := WithRateLimit(time.Second, 2)
	h := mw(http.HandlerFunc(okHandler))

	ip := "10.0.0.2:5678"
	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ip
		h.ServeHTTP(rec, req)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = ip
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
}

func TestWithRateLimit_DifferentIPsAreIndependent(t *testing.T) {
	mw := WithRateLimit(time.Second, 1)
	h := mw(http.HandlerFunc(okHandler))

	for _, ip := range []string{"1.1.1.1:80", "2.2.2.2:80", "3.3.3.3:80"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ip
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("ip %s: expected 200, got %d", ip, rec.Code)
		}
	}
}

func TestRealIP_ForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.5")
	if got := realIP(req); got != "203.0.113.5" {
		t.Fatalf("expected 203.0.113.5, got %s", got)
	}
}

func TestRealIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "198.51.100.1")
	if got := realIP(req); got != "198.51.100.1" {
		t.Fatalf("expected 198.51.100.1, got %s", got)
	}
}

func TestRealIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.1:9000"
	if got := realIP(req); got != "192.0.2.1:9000" {
		t.Fatalf("expected 192.0.2.1:9000, got %s", got)
	}
}
