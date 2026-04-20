package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func okHandlerCORS(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestWithCORS_NilConfig_IsNoOp(t *testing.T) {
	h := WithCORS(nil)(http.HandlerFunc(okHandlerCORS))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	h.ServeHTTP(rec, req)
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("expected no CORS header")
	}
}

func TestWithCORS_EmptyOrigins_IsNoOp(t *testing.T) {
	h := WithCORS(&CORSConfig{})(http.HandlerFunc(okHandlerCORS))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	h.ServeHTTP(rec, req)
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("expected no CORS header")
	}
}

func TestWithCORS_WildcardOrigin_SetsHeader(t *testing.T) {
	cfg := &CORSConfig{AllowedOrigins: []string{"*"}}
	h := WithCORS(cfg)(http.HandlerFunc(okHandlerCORS))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://anything.io")
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://anything.io" {
		t.Fatalf("expected origin header, got %q", got)
	}
}

func TestWithCORS_DisallowedOrigin_NoHeader(t *testing.T) {
	cfg := &CORSConfig{AllowedOrigins: []string{"https://allowed.com"}}
	h := WithCORS(cfg)(http.HandlerFunc(okHandlerCORS))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://notallowed.com")
	h.ServeHTTP(rec, req)
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("expected no CORS header for disallowed origin")
	}
}

func TestWithCORS_Preflight_Returns204(t *testing.T) {
	cfg := &CORSConfig{
		AllowedOrigins: []string{"https://example.com"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         600,
	}
	h := WithCORS(cfg)(http.HandlerFunc(okHandlerCORS))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Max-Age") != "600" {
		t.Fatalf("expected max-age 600, got %s", rec.Header().Get("Access-Control-Max-Age"))
	}
}
