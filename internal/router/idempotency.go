package router

import (
	"net/http"
	"sync"
	"time"
)

// idempotencyEntry holds a cached response for a given idempotency key.
type idempotencyEntry struct {
	statusCode int
	body        []byte
	expiry      time.Time
}

// idempotencyStore is an in-memory store for idempotency keys.
type idempotencyStore struct {
	mu      sync.Mutex
	entries map[string]idempotencyEntry
	ttl     time.Duration
}

func newIdempotencyStore(ttl time.Duration) *idempotencyStore {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &idempotencyStore{
		entries: make(map[string]idempotencyEntry),
		ttl:     ttl,
	}
}

func (s *idempotencyStore) get(key string) (idempotencyEntry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[key]
	if !ok || time.Now().After(e.expiry) {
		delete(s.entries, key)
		return idempotencyEntry{}, false
	}
	return e, true
}

func (s *idempotencyStore) set(key string, statusCode int, body []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[key] = idempotencyEntry{
		statusCode: statusCode,
		body:        body,
		expiry:     time.Now().Add(s.ttl),
	}
}

// WithIdempotency returns middleware that deduplicates requests sharing the
// same Idempotency-Key header value. Repeated requests with the same key
// receive the cached response until ttl expires.
func WithIdempotency(ttl time.Duration) func(http.Handler) http.Handler {
	store := newIdempotencyStore(ttl)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}
			if entry, ok := store.get(key); ok {
				w.Header().Set("X-Idempotent-Replayed", "true")
				w.WriteHeader(entry.statusCode)
				_, _ = w.Write(entry.body)
				return
			}
			rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rec, r)
			store.set(key, rec.statusCode, rec.body)
		})
	}
}

// responseRecorder captures the status code and body written by a handler.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body        []byte
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return r.ResponseWriter.Write(b)
}
