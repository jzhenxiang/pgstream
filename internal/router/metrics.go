package router

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

// requestMetrics holds per-route counters and latency totals.
type requestMetrics struct {
	mu       sync.Mutex
	counts   map[string]int64
	errors   map[string]int64
	totalMs  map[string]int64
}

func newRequestMetrics() *requestMetrics {
	return &requestMetrics{
		counts:  make(map[string]int64),
		errors:  make(map[string]int64),
		totalMs: make(map[string]int64),
	}
}

func (m *requestMetrics) record(path string, statusCode int, elapsed time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counts[path]++
	m.totalMs[path] += elapsed.Milliseconds()
	if statusCode >= 400 {
		m.errors[path]++
	}
}

func (m *requestMetrics) snapshot() map[string]map[string]string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[string]map[string]string, len(m.counts))
	for path, count := range m.counts {
		avg := int64(0)
		if count > 0 {
			avg = m.totalMs[path] / count
		}
		out[path] = map[string]string{
			"requests":    strconv.FormatInt(count, 10),
			"errors":      strconv.FormatInt(m.errors[path], 10),
			"avg_latency_ms": strconv.FormatInt(avg, 10),
		}
	}
	return out
}

// WithMetrics wraps a handler and records per-path request metrics.
func WithMetrics(m *requestMetrics) func(http.Handler) http.Handler {
	if m == nil {
		return func(next http.Handler) http.Handler { return next }
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)
			m.record(r.URL.Path, rw.status, time.Since(start))
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}
