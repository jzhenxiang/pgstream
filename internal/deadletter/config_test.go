package deadletter_test

import (
	"testing"

	"github.com/pgstream/pgstream/internal/deadletter"
)

func TestConfig_Validate_NilConfig(t *testing.T) {
	var c *deadletter.Config
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestConfig_Validate_NegativeCapacity(t *testing.T) {
	c := &deadletter.Config{Capacity: -1}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for negative capacity")
	}
}

func TestConfig_Validate_ZeroCapacity_IsValid(t *testing.T) {
	c := &deadletter.Config{Capacity: 0}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error for zero capacity: %v", err)
	}
}

func TestConfig_Validate_PositiveCapacity(t *testing.T) {
	c := &deadletter.Config{Capacity: 200}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
