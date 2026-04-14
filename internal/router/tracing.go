package router

import (
	"fmt"
	"net/http"
	"time"
)

// TraceEntry records details about a single HTTP request.
type TraceEntry struct {
	Method     string
	Path       string
	StatusCode int
	Duration   time.Duration
	RequestID  string
}

// tracingResponseWriter wraps http.ResponseWriter to capture the status code.
type tracingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *tracingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// WithTracing wraps a handler and records trace entries via the provided sink.
// If sink is nil, the middleware is a no-op.
func WithTracing(sink func(TraceEntry)) func(http.Handler) http.Handler {
	if sink == nil {
		return func(next http.Handler) http.Handler { return next }
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &tracingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r)
			entry := TraceEntry{
				Method:     r.Method,
				Path:       r.URL.Path,
				StatusCode: rw.statusCode,
				Duration:   time.Since(start),
				RequestID:  fmt.Sprintf("%s-%d", r.RemoteAddr, start.UnixNano()),
			}
			sink(entry)
		})
	}
}
