package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/pgstream/internal/middleware"
)

func TestNewSigner_MissingSecret(t *testing.T) {
	_, err := middleware.NewSigner(middleware.SignerConfig{})
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
}

func TestNewSigner_DefaultHeader(t *testing.T) {
	s, err := middleware.NewSigner(middleware.SignerConfig{Secret: "mysecret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil signer")
	}
}

func TestNewSigner_CustomHeader(t *testing.T) {
	s, err := middleware.NewSigner(middleware.SignerConfig{
		Secret: "mysecret",
		Header: "X-Custom-Sig",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil signer")
	}
}

func TestSign_SetsHeader(t *testing.T) {
	s, _ := middleware.NewSigner(middleware.SignerConfig{Secret: "secret"})
	body := []byte(`{"table":"users"}`)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	s.Sign(req, body)

	sig := req.Header.Get("X-PGStream-Signature")
	if sig == "" {
		t.Fatal("expected signature header to be set")
	}
}

func TestVerify_ValidSignature(t *testing.T) {
	s, _ := middleware.NewSigner(middleware.SignerConfig{Secret: "secret"})
	body := []byte(`{"table":"orders"}`)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	s.Sign(req, body)

	sig := req.Header.Get("X-PGStream-Signature")
	err := s.Verify(sig, body, 5*time.Minute)
	if err != nil {
		t.Fatalf("expected valid signature, got: %v", err)
	}
}

func TestVerify_InvalidSignature(t *testing.T) {
	s, _ := middleware.NewSigner(middleware.SignerConfig{Secret: "secret"})
	body := []byte(`{"table":"orders"}`)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	s.Sign(req, body)

	sig := req.Header.Get("X-PGStream-Signature")
	tamperedBody := []byte(`{"table":"accounts"}`)
	err := s.Verify(sig, tamperedBody, 5*time.Minute)
	if err == nil {
		t.Fatal("expected error for tampered body")
	}
}

func TestVerify_MalformedHeader(t *testing.T) {
	s, _ := middleware.NewSigner(middleware.SignerConfig{Secret: "secret"})
	err := s.Verify("bad-header", []byte("body"), 5*time.Minute)
	if err == nil {
		t.Fatal("expected error for malformed header")
	}
}
