package router

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func echoHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func TestWithBodySize_NilConfig_IsNoOp(t *testing.T) {
	h := WithBodySize(nil)(http.HandlerFunc(echoHandler))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("hello"))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithBodySize_ZeroMaxBytes_IsNoOp(t *testing.T) {
	h := WithBodySize(&BodySizeConfig{MaxBytes: 0})(http.HandlerFunc(echoHandler))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("hello"))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithBodySize_BelowLimit_Passes(t *testing.T) {
	cfg := &BodySizeConfig{MaxBytes: 100}
	h := WithBodySize(cfg)(http.HandlerFunc(echoHandler))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(make([]byte, 50)))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithBodySize_ExceedsLimit_ReturnsError(t *testing.T) {
	cfg := &BodySizeConfig{MaxBytes: 10}
	h := WithBodySize(cfg)(http.HandlerFunc(echoHandler))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(make([]byte, 100)))
	h.ServeHTTP(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatal("expected non-200 status when body exceeds limit")
	}
}

func TestDefaultBodySizeConfig_Defaults(t *testing.T) {
	cfg := DefaultBodySizeConfig()
	if cfg.MaxBytes != defaultMaxBodyBytes {
		t.Fatalf("expected %d, got %d", defaultMaxBodyBytes, cfg.MaxBytes)
	}
}

func TestBodySizeConfig_Validate_Nil(t *testing.T) {
	var cfg *BodySizeConfig
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestBodySizeConfig_Validate_NegativeMaxBytes(t *testing.T) {
	cfg := &BodySizeConfig{MaxBytes: -1}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for negative MaxBytes")
	}
}

func TestBodySizeConfig_Validate_ZeroMaxBytes_IsValid(t *testing.T) {
	cfg := &BodySizeConfig{MaxBytes: 0}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestBodySizeConfig_Validate_PositiveMaxBytes(t *testing.T) {
	cfg := &BodySizeConfig{MaxBytes: 512}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
