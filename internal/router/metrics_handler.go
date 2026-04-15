package router

import (
	"encoding/json"
	"net/http"
)

// metricsHandler returns a handler that serialises the current snapshot as JSON.
func metricsHandler(m *requestMetrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		snap := m.snapshot()
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(snap); err != nil {
			http.Error(w, "failed to encode metrics", http.StatusInternalServerError)
		}
	}
}

// RegisterMetrics attaches the metrics snapshot endpoint to the given mux and
// returns the WithMetrics middleware bound to the same store.
func RegisterMetrics(mux *http.ServeMux, cfg *MetricsConfig) (func(http.Handler) http.Handler, error) {
	if cfg == nil || !cfg.Enabled {
		noop := func(next http.Handler) http.Handler { return next }
		return noop, nil
	}
	cfg.setDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	m := newRequestMetrics()
	mux.Handle(cfg.Endpoint, metricsHandler(m))
	return WithMetrics(m), nil
}
