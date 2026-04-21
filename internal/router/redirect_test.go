package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithRedirect_NilConfig_IsNoOp(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := WithRedirect(nil)(handler)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/some/path", nil)
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestWithRedirect_NoRules_PassesThrough(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := WithRedirect(&RedirectConfig{})(handler)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/no/match", nil)
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestWithRedirect_MatchingRule_Redirects(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := &RedirectConfig{
		Rules: []RedirectRule{
			{From: "/old", To: "/new"},
		},
	}

	mw := WithRedirect(cfg)(handler)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/old", nil)
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 301, got %d", rr.Code)
	}

	loc := rr.Header().Get("Location")
	if loc != "/new" {
		t.Fatalf("expected Location /new, got %q", loc)
	}
}

func TestWithRedirect_NonMatchingRule_PassesThrough(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := &RedirectConfig{
		Rules: []RedirectRule{
			{From: "/old", To: "/new"},
		},
	}

	mw := WithRedirect(cfg)(handler)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/other", nil)
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestWithRedirect_CustomCode_UsedInResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := &RedirectConfig{
		Rules: []RedirectRule{
			{From: "/temp", To: "/new-temp", Code: http.StatusTemporaryRedirect},
		},
	}

	mw := WithRedirect(cfg)(handler)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/temp", nil)
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", rr.Code)
	}

	loc := rr.Header().Get("Location")
	if loc != "/new-temp" {
		t.Fatalf("expected Location /new-temp, got %q", loc)
	}
}

func TestWithRedirect_MultipleRules_FirstMatchWins(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := &RedirectConfig{
		Rules: []RedirectRule{
			{From: "/api", To: "/v1/api"},
			{From: "/api", To: "/v2/api"},
		},
	}

	mw := WithRedirect(cfg)(handler)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 301, got %d", rr.Code)
	}

	loc := rr.Header().Get("Location")
	if loc != "/v1/api" {
		t.Fatalf("expected Location /v1/api, got %q", loc)
	}
}
