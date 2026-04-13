package router

import (
	"net/http"
	"runtime/debug"
	"time"
)

// PanicRecord holds details about a recovered panic.
type PanicRecord struct {
	Time       time.Time
	Path       string
	Method     string
	StackTrace string
	Value      any
}

// RecoveryHandler is called when a panic is recovered. It receives the
// request that triggered the panic and a populated PanicRecord.
type RecoveryHandler func(r *http.Request, rec PanicRecord)

// WithRecovery wraps h so that any panic is recovered, the request is
// responded to with 500 Internal Server Error, and the optional handler fn
// is invoked for observability (logging, alerting, etc.).
// If fn is nil the panic is silently swallowed after responding.
func WithRecovery(fn RecoveryHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if v := recover(); v != nil {
					w.WriteHeader(http.StatusInternalServerError)

					if fn != nil {
						fn(r, PanicRecord{
							Time:       time.Now().UTC(),
							Path:       r.URL.Path,
							Method:     r.Method,
							StackTrace: string(debug.Stack()),
							Value:      v,
						})
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
