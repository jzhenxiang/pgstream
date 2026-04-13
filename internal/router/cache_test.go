package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var okHandlerCache = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestWithCache_NilConfig_IsNoOp(t *testing.T) {
	h := WithCache(nil)(okHandlerCache)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := rec.Header().Get("Cache-Control"); got != "" {
		t.Fatalf("expected no Cache-Control header, got %q", got)
	}
}

func TestWithCache_NoStore(t *testing.T) {
	cfg := &CacheConfig{NoStore: true}
	h := WithCache(cfg)(okHandlerCache)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("expected 'no-store', got %q", got)
	}
}

func TestWithCache_PublicMaxAge(t *testing.T) {
	cfg := &CacheConfig{MaxAge: 30 * time.Second}
	h := WithCache(cfg)(okHandlerCache)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=30" {
		t.Fatalf("expected 'public, max-age=30', got %q", got)
	}
}

func TestWithCache_PrivateMaxAge(t *testing.T) {
	cfg := &CacheConfig{Private: true, MaxAge: 120 * time.Second}
	h := WithCache(cfg)(okHandlerCache)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := rec.Header().Get("Cache-Control"); got != "private, max-age=120" {
		t.Fatalf("expected 'private, max-age=120', got %q", got)
	}
}

func TestCacheConfig_Validate_Nil(t *testing.T) {
	var cfg *CacheConfig
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected nil error for nil config, got %v", err)
	}
}

func TestCacheConfig_Validate_NoStoreWithMaxAge_ReturnsError(t *testing.T) {
	cfg := &CacheConfig{NoStore: true, MaxAge: 10 * time.Second}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for NoStore+MaxAge combination, got nil")
	}
}

func TestCacheConfig_Validate_NegativeMaxAge_ReturnsError(t *testing.T) {
	cfg := &CacheConfig{MaxAge: -1 * time.Second}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for negative MaxAge, got nil")
	}
}

func TestCacheConfig_ApplyDefaults_SetsMaxAge(t *testing.T) {
	cfg := &CacheConfig{}
	cfg.ApplyDefaults()
	if cfg.MaxAge != DefaultCacheMaxAge {
		t.Fatalf("expected default MaxAge %v, got %v", DefaultCacheMaxAge, cfg.MaxAge)
	}
}

func TestCacheConfig_ApplyDefaults_NoStoreSkipsMaxAge(t *testing.T) {
	cfg := &CacheConfig{NoStore: true}
	cfg.ApplyDefaults()
	if cfg.MaxAge != 0 {
		t.Fatalf("expected MaxAge to remain 0 when NoStore is set, got %v", cfg.MaxAge)
	}
}
