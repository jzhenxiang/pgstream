package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func bodyHandlerETag(body string, status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	})
}

func TestWithETag_SetsETagHeader(t *testing.T) {
	h := WithETag(bodyHandlerETag("hello", http.StatusOK))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Header().Get("ETag") == "" {
		t.Fatal("expected ETag header to be set")
	}
}

func TestWithETag_MatchingIfNoneMatch_Returns304(t *testing.T) {
	h := WithETag(bodyHandlerETag("hello", http.StatusOK))

	// First request to obtain the ETag.
	rec1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec1, req1)
	etag := rec1.Header().Get("ETag")

	// Second request with If-None-Match.
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("If-None-Match", etag)
	h.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusNotModified {
		t.Fatalf("expected 304, got %d", rec2.Code)
	}
}

func TestWithETag_NonGetMethod_SkipsETag(t *testing.T) {
	h := WithETag(bodyHandlerETag("ok", http.StatusOK))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Header().Get("ETag") != "" {
		t.Fatal("expected no ETag header for POST request")
	}
}

func TestWithETag_NonOKStatus_SkipsETag(t *testing.T) {
	h := WithETag(bodyHandlerETag("not found", http.StatusNotFound))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Header().Get("ETag") != "" {
		t.Fatal("expected no ETag header for non-200 response")
	}
}

func TestWithETag_DifferentBodies_DifferentETags(t *testing.T) {
	h1 := WithETag(bodyHandlerETag("body-one", http.StatusOK))
	h2 := WithETag(bodyHandlerETag("body-two", http.StatusOK))

	rec1 := httptest.NewRecorder()
	h1.ServeHTTP(rec1, httptest.NewRequest(http.MethodGet, "/", nil))

	rec2 := httptest.NewRecorder()
	h2.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec1.Header().Get("ETag") == rec2.Header().Get("ETag") {
		t.Fatal("expected different ETags for different bodies")
	}
}
