package ratelimit

import (
	"testing"
)

func TestConfig_IsDisabled_ZeroRate(t *testing.T) {
	cfg := Config{Rate: 0}
	if !cfg.disabled() {
		t.Fatal("expected zero rate to be disabled")
	}
}

func TestConfig_IsDisabled_PositiveRate(t *testing.T) {
	cfg := Config{Rate: 100}
	if cfg.disabled() {
		t.Fatal("expected positive rate to not be disabled")
	}
}

func TestConfig_IsDisabled_NegativeRate(t *testing.T) {
	cfg := Config{Rate: -1}
	// negative rates are rejected at construction time, but the helper
	// itself only checks for zero
	if cfg.disabled() {
		t.Fatal("expected negative rate to not be considered disabled by helper")
	}
}

func TestConfig_BurstDefault(t *testing.T) {
	cfg := Config{Rate: 50}
	if cfg.burst() != 50 {
		t.Fatalf("expected burst to equal rate 50, got %d", cfg.burst())
	}
}

func TestConfig_BurstExplicit(t *testing.T) {
	cfg := Config{Rate: 50, Burst: 200}
	if cfg.burst() != 200 {
		t.Fatalf("expected burst 200, got %d", cfg.burst())
	}
}
