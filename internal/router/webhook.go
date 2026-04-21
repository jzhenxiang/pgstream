package router

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pgstream/pgstream/internal/wal"
)

// WebhookPayload is the request body sent to a registered webhook endpoint.
type WebhookPayload struct {
	Event     *wal.Event `json:"event"`
	Timestamp time.Time  `json:"timestamp"`
	Source    string     `json:"source,omitempty"`
}

// WebhookHandler returns an http.Handler that accepts incoming WAL events
// encoded as JSON and forwards them to the provided sink function.
func WebhookHandler(sink func(*wal.Event) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload WebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if payload.Event == nil {
			http.Error(w, "missing event", http.StatusBadRequest)
			return
		}

		if err := sink(payload.Event); err != nil {
			http.Error(w, "failed to process event", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}
