package router

import "errors"

// MetricsConfig controls the /metrics endpoint and collection behaviour.
type MetricsConfig struct {
	// Enabled controls whether the middleware collects metrics.
	Enabled bool
	// Endpoint is the HTTP path that exposes a snapshot (default: /metrics).
	Endpoint string
}

func (c *MetricsConfig) setDefaults() {
	if c.Endpoint == "" {
		c.Endpoint = "/metrics"
	}
}

// Validate returns an error when the config is invalid.
func (c *MetricsConfig) Validate() error {
	if c == nil {
		return errors.New("metrics config must not be nil")
	}
	if c.Endpoint == "" {
		return errors.New("metrics endpoint must not be blank")
	}
	return nil
}
