package sink

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookConfig holds configuration for the webhook sink.
type WebhookConfig struct {
	URL     string
	Timeout time.Duration
	Headers map[string]string
}

// WebhookSink delivers WAL events to an.
type WebhookSink struct {
	client *http.Client
}

// NewWebhookSink creates a Web.
func NewWebhookSink(cfg WebhookConfig) (*WebhookSink, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("webhook: URL must not be empty")
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &WebhookSink{
		cfg:    cfg,
		client: &http.Client{Timeout: timeout},
	}, nil
}

// Send marshals the event to JSON and POSTs it to the configured URL.
func (w *WebhookSink) Send(ctx context.Context, event Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("webhook: marshal event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.cfg.URL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("webhook: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range w.cfg.Headers {
		req.Header.Set(k, v)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d", resp.StatusCode)
	}
	return nil
}

// Close is a no-op for the webhook sink.
func (w *WebhookSink) Close() error { return nil }
