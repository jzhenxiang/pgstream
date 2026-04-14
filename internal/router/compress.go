package router

import (
	"compress/gzip"
	"net/http"
	"strings"
	"sync"
)

var gzipPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(nil)
	},
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.writer.Write(b)
}

// WithCompress returns a middleware that gzip-compresses responses when the
// client advertises Accept-Encoding: gzip. Responses smaller than minBytes are
// passed through uncompressed.
func WithCompress(minBytes int) func(http.Handler) http.Handler {
	if minBytes <= 0 {
		minBytes = 512
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			gz := gzipPool.Get().(*gzip.Writer)
			gz.Reset(w)
			defer func() {
				_ = gz.Close()
				gzipPool.Put(gz)
			}()

			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Del("Content-Length")

			grw := &gzipResponseWriter{ResponseWriter: w, writer: gz}
			next.ServeHTTP(grw, r)
		})
	}
}
