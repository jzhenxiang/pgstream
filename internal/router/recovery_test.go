package router

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func panicHandler(w http.ResponseWriter, _ *http.Request) {
	panic("boom")
}

func okHandlerRecovery(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestWithRecovery_NilFn_IsNoOp(t *testing.T) {
	h := WithRecovery(nil)(http.HandlerFunc(okHandlerRecovery))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithRecovery_PanicReturns500(t *testing.T) {
	h := WithRecovery(nil)(http.HandlerFunc(panicHandler))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/crash", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestWithRecovery_HandlerCalled(t *testing.T) {
	var called atomic.Bool
	fn := func(_ *http.Request, rec PanicRecord) {
		called.Store(true)
		if rec.Value != "boom" {
			t.Errorf("unexpected panic value: %v", rec.Value)
		}
		if rec.Path != "/crash" {
			t.Errorf("unexpected path: %s", rec.Path)
		}
		if rec.StackTrace == "" {
			t.Error("expected non-empty stack trace")
		}
	}

	h := WithRecovery(fn)(http.HandlerFunc(panicHandler))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/crash", nil))

	if !called.Load() {
		t.Fatal("recovery handler was not called")
	}
}

func TestWithRecovery_NoPanic_HandlerNotCalled(t *testing.T) {
	var called atomic.Bool
	fn := func(_ *http.Request, _ PanicRecord) { called.Store(true) }

	h := WithRecovery(fn)(http.HandlerFunc(okHandlerRecovery))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if called.Load() {
		t.Fatal("recovery handler should not have been called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
