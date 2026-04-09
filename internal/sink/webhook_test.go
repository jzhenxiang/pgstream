package sink

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewWebhookSink_MissingURL(t *testing.T) {
	_, err := NewWebhookSink(WebhookConfig{})
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
}

func TestWebhookSink_Send_Success(t *testing.T) {
	var received Event
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("unexpected Content-Type: %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sink, err := NewWebhookSink(WebhookConfig{URL: server.URL, Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("NewWebhookSink: %v", err)
	}

	event := Event{LSN: "0/1234", Table: "orders", Action: "INSERT", Data: map[string]interface{}{"id": 1}}
	if err := sink.Send(context.Background(), event); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if received.Table != "orders" {
		t.Errorf("expected table 'orders', got %q", received.Table)
	}
}

func TestWebhookSink_Send_NonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	sink, _ := NewWebhookSink(WebhookConfig{URL: server.URL})
	err := sink.Send(context.Background(), Event{Table: "test", Action: "DELETE"})
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}
