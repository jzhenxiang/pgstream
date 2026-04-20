package router

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
)

// etagResponseWriter wraps http.ResponseWriter to capture the response body
// so an ETag can be computed from it.
type etagResponseWriter struct {
	http.ResponseWriter
	body   []byte
	status int
}

func (e *etagResponseWriter) WriteHeader(status int) {
	e.status = status
	e.ResponseWriter.WriteHeader(status)
}

func (e *etagResponseWriter) Write(b []byte) (int, error) {
	e.body = append(e.body, b...)
	return e.ResponseWriter.Write(b)
}

// WithETag adds ETag support to the router. It computes a SHA-256 based ETag
// from the response body and sets the ETag header. If the client sends a
// matching If-None-Match header the middleware short-circuits with 304.
func WithETag(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only apply ETag logic to safe methods.
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			next.ServeHTTP(w, r)
			return
		}

		erw := &etagResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(erw, r)

		// Only set ETag on successful responses.
		if erw.status != http.StatusOK {
			return
		}

		etag := computeETag(erw.body)
		w.Header().Set("ETag", etag)

		if match := r.Header.Get("If-None-Match"); match != "" {
			if strings.TrimSpace(match) == etag {
				w.WriteHeader(http.StatusNotModified)
			}
		}
	})
}

func computeETag(body []byte) string {
	sum := sha256.Sum256(body)
	return fmt.Sprintf(`"%x"`, sum[:8])
}
