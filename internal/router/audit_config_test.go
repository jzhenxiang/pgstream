package router

import "testing"

func TestAuditConfig_Validate_Nil(t *testing.T) {
	var c *AuditConfig
	if err := c.Validate(); err == nil {
		t.Error("expected error for nil config")
	}
}

func TestAuditConfig_Validate_NegativeMaxEntries(t *testing.T) {
	c := &AuditConfig{MaxEntries: -1}
	if err := c.Validate(); err == nil {
		t.Error("expected error for negative MaxEntries")
	}
}

func TestAuditConfig_Validate_ZeroMaxEntries_IsValid(t *testing.T) {
	c := &AuditConfig{MaxEntries: 0}
	if err := c.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAuditConfig_Validate_PositiveMaxEntries(t *testing.T) {
	c := &AuditConfig{MaxEntries: 500}
	if err := c.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDefaultAuditConfig_Defaults(t *testing.T) {
	c := DefaultAuditConfig()
	if c.MaxEntries != 1000 {
		t.Errorf("expected MaxEntries=1000, got %d", c.MaxEntries)
	}
}

func TestNewInMemoryAuditSink_ZeroMax_UsesDefault(t *testing.T) {
	sink := NewInMemoryAuditSink(0).(*inMemoryAuditSink)
	if sink.max != DefaultAuditConfig().MaxEntries {
		t.Errorf("expected default max %d, got %d", DefaultAuditConfig().MaxEntries, sink.max)
	}
}
