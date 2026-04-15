package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func slowHandler(d time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(d):
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
		}
	})
}

func TestWithRequestTimeout_NilConfig_PassesThrough(t *testing.T) {
	h := WithRequestTimeout(nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithRequestTimeout_ZeroDuration_PassesThrough(t *testing.T) {
	cfg := &TimeoutConfig{Duration: 0}
	h := WithRequestTimeout(cfg)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithRequestTimeout_FastHandler_Returns200(t *testing.T) {
	cfg := &TimeoutConfig{Duration: 500 * time.Millisecond}
	h := WithRequestTimeout(cfg)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithRequestTimeout_SlowHandler_Returns503(t *testing.T) {
	cfg := &TimeoutConfig{Duration: 50 * time.Millisecond, Message: "timed out"}
	h := WithRequestTimeout(cfg)(slowHandler(200 * time.Millisecond))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestDefaultTimeoutConfig_Defaults(t *testing.T) {
	cfg := DefaultTimeoutConfig()
	if cfg.Duration != DefaultTimeoutDuration {
		t.Fatalf("expected %v, got %v", DefaultTimeoutDuration, cfg.Duration)
	}
	if cfg.Message == "" {
		t.Fatal("expected non-empty message")
	}
}

func TestTimeoutConfig_Validate_NegativeDuration(t *testing.T) {
	cfg := &TimeoutConfig{Duration: -1 * time.Second}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for negative duration")
	}
}

func TestTimeoutConfig_Validate_Valid(t *testing.T) {
	cfg := &TimeoutConfig{Duration: 10 * time.Second}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTimeoutConfig_Validate_Nil(t *testing.T) {
	var cfg *TimeoutConfig
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error for nil config: %v", err)
	}
}
