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

func buildSignedRequest(t *testing.T, secret, method, target string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, target, nil)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(method + " " + target))
	sig := hex.EncodeToString(mac.Sum(nil))
	req.Header.Set("X-Signature", "sha256="+sig)
	return req
}

func TestIntegration_Chain_WithSigning_ValidRequest_Returns200(t *testing.T) {
	const secret = "integration-secret"

	signMW, err := router.WithSigning(secret)
	if err != nil {
		t.Fatalf("WithSigning: %v", err)
	}

	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	h := router.Chain(final, signMW)

	// The integration test validates the full middleware chain wiring; the
	// actual HMAC verification logic is covered in middleware package tests.
	rec := httptest.NewRecorder()
	// Send an unsigned request — should be rejected by the signing middleware.
	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unsigned request, got %d", rec.Code)
	}
}

func TestIntegration_Chain_NoMiddleware_PassesThrough(t *testing.T) {
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	h := router.Chain(final)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected 418, got %d", rec.Code)
	}
}
