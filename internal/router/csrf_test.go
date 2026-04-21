package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func okHandlerCSRF(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestWithCSRF_NilConfig_IsNoOp(t *testing.T) {
	mw := WithCSRF(nil)
	handler := mw(http.HandlerFunc(okHandlerCSRF))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestWithCSRF_EmptySecret_IsNoOp(t *testing.T) {
	mw := WithCSRF(&CSRFConfig{Secret: ""})
	handler := mw(http.HandlerFunc(okHandlerCSRF))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestWithCSRF_SafeMethod_SetsCookie(t *testing.T) {
	mw := WithCSRF(&CSRFConfig{Secret: "supersecret"})
	handler := mw(http.HandlerFunc(okHandlerCSRF))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	cookies := rr.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected csrf cookie to be set")
	}
	var found bool
	for _, c := range cookies {
		if c.Name == "csrf_token" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("csrf_token cookie not found")
	}
}

func TestWithCSRF_UnsafeMethod_MissingCookie_Returns403(t *testing.T) {
	mw := WithCSRF(&CSRFConfig{Secret: "supersecret"})
	handler := mw(http.HandlerFunc(okHandlerCSRF))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestWithCSRF_UnsafeMethod_MissingHeader_Returns403(t *testing.T) {
	mw := WithCSRF(&CSRFConfig{Secret: "supersecret"})
	handler := mw(http.HandlerFunc(okHandlerCSRF))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	token := generateCSRFToken("supersecret")
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: token})
	// header intentionally omitted
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestWithCSRF_UnsafeMethod_ValidToken_Returns200(t *testing.T) {
	mw := WithCSRF(&CSRFConfig{Secret: "supersecret"})
	handler := mw(http.HandlerFunc(okHandlerCSRF))

	token := generateCSRFToken("supersecret")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: token})
	req.Header.Set("X-CSRF-Token", token)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestWithCSRF_CustomHeader_And_Cookie(t *testing.T) {
	cfg := &CSRFConfig{
		Secret:      "mysecret",
		TokenHeader: "X-My-CSRF",
		CookieName:  "my_csrf",
	}
	mw := WithCSRF(cfg)
	handler := mw(http.HandlerFunc(okHandlerCSRF))

	token := generateCSRFToken("mysecret")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/resource", nil)
	req.AddCookie(&http.Cookie{Name: "my_csrf", Value: token})
	req.Header.Set("X-My-CSRF", token)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
