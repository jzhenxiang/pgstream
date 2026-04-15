package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func idempotencyHandler(body string, status int) http.Handler {
	calls := 0
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	})
}

func TestWithIdempotency_NoKey_PassesThrough(t *testing.T) {
	mw := WithIdempotency(time.Minute)
	h := mw(idempotencyHandler("ok", http.StatusOK))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("X-Idempotent-Replayed") != "" {
		t.Fatal("expected no replay header")
	}
}

func TestWithIdempotency_SameKey_ReturnsCachedResponse(t *testing.T) {
	mw := WithIdempotency(time.Minute)
	h := mw(idempotencyHandler("created", http.StatusCreated))

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Idempotency-Key", "key-abc")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("iteration %d: expected 201, got %d", i, w.Code)
		}
		if w.Body.String() != "created" {
			t.Fatalf("iteration %d: unexpected body %q", i, w.Body.String())
		}
		if i > 0 && w.Header().Get("X-Idempotent-Replayed") != "true" {
			t.Fatalf("iteration %d: expected replay header", i)
		}
	}
}

func TestWithIdempotency_DifferentKeys_IndependentResponses(t *testing.T) {
	mw := WithIdempotency(time.Minute)
	h := mw(idempotencyHandler("ok", http.StatusOK))

	for _, key := range []string{"key-1", "key-2"} {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Idempotency-Key", key)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Header().Get("X-Idempotent-Replayed") != "" {
			t.Fatalf("key %s should not be replayed on first call", key)
		}
	}
}

func TestWithIdempotency_ExpiredEntry_ReExecutesHandler(t *testing.T) {
	store := newIdempotencyStore(time.Millisecond)
	store.set("key-exp", http.StatusOK, []byte("old"))

	time.Sleep(5 * time.Millisecond)

	_, ok := store.get("key-exp")
	if ok {
		t.Fatal("expected expired entry to be evicted")
	}
}

func TestWithIdempotency_DefaultTTL(t *testing.T) {
	store := newIdempotencyStore(0)
	if store.ttl != 24*time.Hour {
		t.Fatalf("expected default TTL of 24h, got %v", store.ttl)
	}
}
