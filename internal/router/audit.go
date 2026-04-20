package router

import (
	"net/http"
	"time"
)

// AuditEntry records a single HTTP request for audit purposes.
type AuditEntry struct {
	Timestamp  time.Time
	Method     string
	Path       string
	StatusCode int
	RequestID  string
	RemoteAddr string
	DurationMs int64
}

// AuditSink receives audit entries.
type AuditSink interface {
	Record(entry AuditEntry)
}

// WithAudit wraps a handler and records an AuditEntry for every request.
// If sink is nil the middleware is a no-op.
func WithAudit(sink AuditSink) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if sink == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &auditResponseWriter{ResponseWriter: w, code: http.StatusOK}
			next.ServeHTTP(rw, r)
			sink.Record(AuditEntry{
				Timestamp:  start,
				Method:     r.Method,
				Path:       r.URL.Path,
				StatusCode: rw.code,
				RequestID:  RequestIDFromContext(r.Context()),
				RemoteAddr: realIP(r),
				DurationMs: time.Since(start).Milliseconds(),
			})
		})
	}
}

type auditResponseWriter struct {
	http.ResponseWriter
	code int
}

func (a *auditResponseWriter) WriteHeader(code int) {
	a.code = code
	a.ResponseWriter.WriteHeader(code)
}
