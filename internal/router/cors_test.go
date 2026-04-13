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
	mw := WithCORS(nil)
	h := mw(http.HandlerFunc(okHandlerCORS))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	h.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("expected no CORS header for nil config")
	}
}

func TestWithCORS_EmptyOrigins_IsNoOp(t *testing.T) {
	mw := WithCORS(&CORSConfig{})
	h := mw(http.HandlerFunc(okHandlerCORS))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	h.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("expected no CORS header for empty origins")
	}
}

func TestWithCORS_WildcardOrigin_SetsHeader(t *testing.T) {
	mw := WithCORS(&CORSConfig{AllowedOrigins: []string{"*"}})
	h := mw(http.HandlerFunc(okHandlerCORS))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://any.com")
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://any.com" {
		t.Fatalf("expected origin header, got %q", got)
	}
}

func TestWithCORS_DisallowedOrigin_NoHeader(t *testing.T) {
	mw := WithCORS(&CORSConfig{AllowedOrigins: []string{"https://trusted.com"}})
	h := mw(http.HandlerFunc(okHandlerCORS))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.com")
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no CORS header for disallowed origin, got %q", got)
	}
}

func TestWithCORS_PreflightOptions_Returns204(t *testing.T) {
	mw := WithCORS(&CORSConfig{
		AllowedOrigins: []string{"https://trusted.com"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         3600,
	})
	h := mw(http.HandlerFunc(okHandlerCORS))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://trusted.com")
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Max-Age") != "3600" {
		t.Fatal("expected Max-Age header")
	}
}

func TestWithCORS_DefaultMethods_Applied(t *testing.T) {
	mw := WithCORS(&CORSConfig{AllowedOrigins: []string{"*"}})
	h := mw(http.HandlerFunc(okHandlerCORS))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, OPTIONS" {
		t.Fatalf("expected default methods, got %q", got)
	}
}
