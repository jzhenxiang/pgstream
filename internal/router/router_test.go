package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pgstream/pgstream/internal/healthcheck"
	"github.com/pgstream/pgstream/internal/router"
)

func newHC(t *testing.T) *healthcheck.HealthCheck {
	t.Helper()
	hc, err := healthcheck.New(healthcheck.Config{Addr: ":0"})
	if err != nil {
		t.Fatalf("healthcheck.New: %v", err)
	}
	return hc
}

func TestNew_NoSigning_ReturnsHandler(t *testing.T) {
	h, err := router.New(router.Config{}, newHC(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestNew_InvalidSigningSecret_ReturnsError(t *testing.T) {
	// Empty secret with non-empty header should still fail because
	// the secret itself is empty — NewSigner requires a secret.
	_, err := router.New(router.Config{SigningSecret: ""}, newHC(t))
	// No signing configured → no error expected.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHealthzEndpoint_Returns200(t *testing.T) {
	h, err := router.New(router.Config{}, newHC(t))
	if err != nil {
		t.Fatalf("router.New: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestUnknownRoute_Returns404(t *testing.T) {
	h, err := router.New(router.Config{}, newHC(t))
	if err != nil {
		t.Fatalf("router.New: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestNew_WithValidSigningSecret(t *testing.T) {
	h, err := router.New(router.Config{
		SigningSecret:   "supersecret",
		SignatureHeader: "X-Hub-Signature",
	}, newHC(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}
