package router

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pgstream/pgstream/internal/wal"
)

func TestWebhookHandler_MethodNotAllowed(t *testing.T) {
	h := WebhookHandler(func(_ *wal.Event) error { return nil })
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestWebhookHandler_InvalidBody(t *testing.T) {
	h := WebhookHandler(func(_ *wal.Event) error { return nil })
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("not-json"))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestWebhookHandler_MissingEvent(t *testing.T) {
	h := WebhookHandler(func(_ *wal.Event) error { return nil })
	body, _ := json.Marshal(WebhookPayload{Timestamp: time.Now()})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestWebhookHandler_SinkError(t *testing.T) {
	h := WebhookHandler(func(_ *wal.Event) error { return errors.New("sink down") })
	body, _ := json.Marshal(WebhookPayload{Event: &wal.Event{}, Timestamp: time.Now()})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestWebhookHandler_Success(t *testing.T) {
	var received *wal.Event
	h := WebhookHandler(func(e *wal.Event) error {
		received = e
		return nil
	})
	event := &wal.Event{}
	body, _ := json.Marshal(WebhookPayload{Event: event, Timestamp: time.Now()})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if received == nil {
		t.Fatal("expected sink to receive event")
	}
}
