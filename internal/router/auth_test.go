package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func okHandlerAuth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestWithAuth_NilConfig_IsNoOp(t *testing.T) {
	h := WithAuth(nil)(http.HandlerFunc(okHandlerAuth))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithAuth_EmptyTokens_IsNoOp(t *testing.T) {
	cfg := &AuthConfig{Tokens: []string{}}
	h := WithAuth(cfg)(http.HandlerFunc(okHandlerAuth))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 (no-op), got %d", rec.Code)
	}
}

func TestWithAuth_MissingHeader_Returns401(t *testing.T) {
	cfg := &AuthConfig{Tokens: []string{"secret"}}
	h := WithAuth(cfg)(http.HandlerFunc(okHandlerAuth))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if rec.Header().Get("WWW-Authenticate") == "" {
		t.Fatal("expected WWW-Authenticate header to be set")
	}
}

func TestWithAuth_WrongToken_Returns401(t *testing.T) {
	cfg := &AuthConfig{Tokens: []string{"secret"}}
	h := WithAuth(cfg)(http.HandlerFunc(okHandlerAuth))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestWithAuth_ValidToken_Returns200(t *testing.T) {
	cfg := &AuthConfig{Tokens: []string{"secret"}}
	h := WithAuth(cfg)(http.HandlerFunc(okHandlerAuth))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithAuth_DefaultRealm(t *testing.T) {
	cfg := &AuthConfig{Tokens: []string{"tok"}}
	h := WithAuth(cfg)(http.HandlerFunc(okHandlerAuth))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	www := rec.Header().Get("WWW-Authenticate")
	if www != `Bearer realm="pgstream"` {
		t.Fatalf("unexpected WWW-Authenticate: %q", www)
	}
}

func TestWithAuth_CustomRealm(t *testing.T) {
	cfg := &AuthConfig{Tokens: []string{"tok"}, Realm: "myapp"}
	h := WithAuth(cfg)(http.HandlerFunc(okHandlerAuth))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	www := rec.Header().Get("WWW-Authenticate")
	if www != `Bearer realm="myapp"` {
		t.Fatalf("unexpected WWW-Authenticate: %q", www)
	}
}
