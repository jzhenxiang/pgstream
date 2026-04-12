package jitter

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Factor != 0.2 {
		t.Fatalf("expected default factor 0.2, got %v", cfg.Factor)
	}
}

func TestNew_ZeroFactor_UsesDefault(t *testing.T) {
	j, err := New(Config{Factor: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if j.cfg.Factor != DefaultConfig().Factor {
		t.Fatalf("expected default factor, got %v", j.cfg.Factor)
	}
}

func TestNew_NegativeFactor_UsesDefault(t *testing.T) {
	j, err := New(Config{Factor: -0.5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if j.cfg.Factor != DefaultConfig().Factor {
		t.Fatalf("expected default factor, got %v", j.cfg.Factor)
	}
}

func TestNew_ValidFactor(t *testing.T) {
	j, err := New(Config{Factor: 0.5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if j.cfg.Factor != 0.5 {
		t.Fatalf("expected factor 0.5, got %v", j.cfg.Factor)
	}
}

func TestApply_ZeroBase_ReturnsZero(t *testing.T) {
	j, _ := New(DefaultConfig())
	if got := j.Apply(0); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestApply_AddsPositiveJitter(t *testing.T) {
	base := 100 * time.Millisecond
	j, _ := New(Config{Factor: 0.2})
	for i := 0; i < 50; i++ {
		got := j.Apply(base)
		if got < base {
			t.Fatalf("Apply should never return less than base: got %v", got)
		}
		max := base + time.Duration(float64(base)*0.2)
		if got > max {
			t.Fatalf("Apply exceeded max jitter: got %v, max %v", got, max)
		}
	}
}

func TestApplyFull_StaysWithinBounds(t *testing.T) {
	base := 200 * time.Millisecond
	j, _ := New(Config{Factor: 0.1})
	spread := time.Duration(float64(base) * 0.1)
	lo := base - spread
	hi := base + spread
	for i := 0; i < 50; i++ {
		got := j.ApplyFull(base)
		if got < lo || got > hi {
			t.Fatalf("ApplyFull out of bounds: got %v, want [%v, %v]", got, lo, hi)
		}
	}
}

func TestApplyFull_NegativeBase_ReturnsBase(t *testing.T) {
	j, _ := New(DefaultConfig())
	base := -1 * time.Second
	if got := j.ApplyFull(base); got != base {
		t.Fatalf("expected base returned for negative input, got %v", got)
	}
}
