package router_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pgstream/pgstream/internal/router"
)

const testSecret = "integration-secret"

func signedRequest(t *testing.T, method, target, header string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, target, nil)
	mac := hmac.New(sha256.New, []byte(testSecret))
	mac.Write([]byte("")) // empty body
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if header == "" {
		header = "X-Signature"
	}
	req.Header.Set(header, sig)
	return req
}

func TestIntegration_SignedHealthz_Returns200(t *testing.T) {
	h, err := router.New(router.Config{
		SigningSecret: testSecret,
	}, newHC(t))
	if err != nil {
		t.Fatalf("router.New: %v", err)
	}

	rec := httptest.NewRecorder()
	req := signedRequest(t, http.MethodGet, "/healthz", "")
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestIntegration_UnsignedRequest_Returns401(t *testing.T) {
	h, err := router.New(router.Config{
		SigningSecret: testSecret,
	}, newHC(t))
	if err != nil {
		t.Fatalf("router.New: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil) // no signature
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestIntegration_WrongSecret_Returns401(t *testing.T) {
	h, err := router.New(router.Config{
		SigningSecret: testSecret,
	}, newHC(t))
	if err != nil {
		t.Fatalf("router.New: %v", err)
	}

	// Build a signature using a different secret so it won't match.
	mac := hmac.New(sha256.New, []byte("wrong-secret"))
	mac.Write([]byte(""))
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("X-Signature", sig)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
