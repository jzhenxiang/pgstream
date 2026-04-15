package router

import "testing"

func TestMetricsConfig_Validate_Nil(t *testing.T) {
	var c *MetricsConfig
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestMetricsConfig_Validate_BlankEndpoint(t *testing.T) {
	c := &MetricsConfig{Enabled: true, Endpoint: ""}
	c.setDefaults()
	if err := c.Validate(); err != nil {
		t.Fatalf("expected no error after setDefaults, got %v", err)
	}
}

func TestMetricsConfig_Validate_Valid(t *testing.T) {
	c := &MetricsConfig{Enabled: true, Endpoint: "/metrics"}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegisterMetrics_NilConfig_ReturnsNoop(t *testing.T) {
	mux := newTestMux()
	_, err := RegisterMetrics(mux, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegisterMetrics_DisabledConfig_ReturnsNoop(t *testing.T) {
	mux := newTestMux()
	_, err := RegisterMetrics(mux, &MetricsConfig{Enabled: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegisterMetrics_Enabled_RegistersEndpoint(t *testing.T) {
	mux := newTestMux()
	_, err := RegisterMetrics(mux, &MetricsConfig{Enabled: true, Endpoint: "/metrics"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func newTestMux() *http.ServeMux {
	return http.NewServeMux()
}
