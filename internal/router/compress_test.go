package router

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func bodyHandler(body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, body)
	})
}

func TestWithCompress_NoAcceptEncoding_PassesThrough(t *testing.T) {
	h := WithCompress(0)(bodyHandler("hello"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") == "gzip" {
		t.Fatal("expected no gzip encoding")
	}
	if rec.Body.String() != "hello" {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestWithCompress_AcceptGzip_CompressesResponse(t *testing.T) {
	h := WithCompress(0)(bodyHandler("hello world"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	h.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatal("expected gzip content-encoding")
	}

	gr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("gzip.NewReader: %v", err)
	}
	defer gr.Close()

	got, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("read gzip body: %v", err)
	}
	if string(got) != "hello world" {
		t.Fatalf("unexpected decompressed body: %s", got)
	}
}

func TestWithCompress_ContentLengthRemoved(t *testing.T) {
	h := WithCompress(512)(bodyHandler(strings.Repeat("x", 1024)))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	h.ServeHTTP(rec, req)

	if cl := rec.Header().Get("Content-Length"); cl != "" {
		t.Fatalf("expected Content-Length to be removed, got %s", cl)
	}
}

func TestWithCompress_ZeroMinBytes_UsesDefault(t *testing.T) {
	// zero minBytes should not panic and should still compress
	h := WithCompress(0)(bodyHandler("data"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
}
