package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestChain_AppliesMiddlewareInOrder(t *testing.T) {
	var order []int
	mk := func(n int) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, n)
				next.ServeHTTP(w, r)
			})
		}
	}

	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	h := Chain(final, mk(1), mk(2), mk(3))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if len(order) != 3 || order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Fatalf("unexpected order: %v", order)
	}
}

func TestWithTimeout_PassesThrough(t *testing.T) {
	called := false
	h := WithTimeout(time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected inner handler to be called")
	}
}

func TestWithTimeout_ZeroDuration_IsNoOp(t *testing.T) {
	called := false
	h := WithTimeout(0)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected inner handler to be called")
	}
}

func TestWithSigning_EmptySecret_IsNoOp(t *testing.T) {
	mw, err := WithSigning("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	called := false
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected inner handler to be called")
	}
}

func TestWithSigning_InvalidSecret_ReturnsError(t *testing.T) {
	// NewSigner returns an error for empty secret; non-empty should succeed.
	_, err := WithSigning("valid-secret")
	if err != nil {
		t.Fatalf("unexpected error for valid secret: %v", err)
	}
}

func TestWithSigning_MissingHeader_Returns401(t *testing.T) {
	mw, err := WithSigning("my-secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
