package router

import "testing"

func TestTracingConfig_Validate_Nil(t *testing.T) {
	var c *TracingConfig
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestTracingConfig_Validate_NegativeMaxEntries(t *testing.T) {
	c := &TracingConfig{Enabled: true, MaxEntries: -1}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for negative MaxEntries")
	}
}

func TestTracingConfig_Validate_ZeroMaxEntries_IsValid(t *testing.T) {
	c := &TracingConfig{Enabled: true, MaxEntries: 0}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTracingConfig_Validate_Valid(t *testing.T) {
	c := DefaultTracingConfig()
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDefaultTracingConfig_Defaults(t *testing.T) {
	c := DefaultTracingConfig()
	if !c.Enabled {
		t.Error("expected Enabled to be true")
	}
	if c.MaxEntries != 1000 {
		t.Errorf("expected MaxEntries 1000, got %d", c.MaxEntries)
	}
}
