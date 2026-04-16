package router

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthzConfig holds configuration for the health endpoint.
type HealthzConfig struct {
	// Endpoint is the path to register the health handler on.
	// Defaults to "/healthz".
	Endpoint string
	// Checks is an optional map of named readiness checks.
	Checks map[string]func() error
}

type healthzResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
	Time   time.Time         `json:"time"`
}

// WithHealthz registers a health check endpoint on the given mux.
// If cfg is nil or Endpoint is empty, "/healthz" is used.
func WithHealthz(mux *http.ServeMux, cfg *HealthzConfig) {
	endpoint := "/healthz"
	var checks map[string]func() error

	if cfg != nil {
		if cfg.Endpoint != "" {
			endpoint = cfg.Endpoint
		}
		checks = cfg.Checks
	}

	mux.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		resp := healthzResponse{
			Status: "ok",
			Time:   time.Now().UTC(),
		}

		if len(checks) > 0 {
			resp.Checks = make(map[string]string, len(checks))
			for name, fn := range checks {
				if err := fn(); err != nil {
					resp.Checks[name] = err.Error()
					resp.Status = "degraded"
				} else {
					resp.Checks[name] = "ok"
				}
			}
		}

		code := http.StatusOK
		if resp.Status != "ok" {
			code = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(resp)
	})
}
