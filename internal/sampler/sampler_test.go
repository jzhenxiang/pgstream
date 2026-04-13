package sampler

import (
	"testing"

	"github.com/pgstream/pgstream/internal/wal"
)

func TestNew_InvalidRate_Zero(t *testing.T) {
	_, err := New(Config{Rate: 0.0})
	if err == nil {
		t.Fatal("expected error for zero rate, got nil")
	}
}

func TestNew_InvalidRate_Negative(t *testing.T) {
	_, err := New(Config{Rate: -0.5})
	if err == nil {
		t.Fatal("expected error for negative rate, got nil")
	}
}

func TestNew_InvalidRate_AboveOne(t *testing.T) {
	_, err := New(Config{Rate: 1.1})
	if err == nil {
		t.Fatal("expected error for rate > 1.0, got nil")
	}
}

func TestNew_Valid(t *testing.T) {
	s, err := New(Config{Rate: 0.5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil sampler")
	}
}

func TestSample_NilEvent_ReturnsFalse(t *testing.T) {
	s, _ := New(Config{Rate: 1.0})
	if s.Sample(nil) {
		t.Fatal("expected false for nil event")
	}
}

func TestSample_RateOne_AlwaysTrue(t *testing.T) {
	s, _ := New(Config{Rate: 1.0})
	event := &wal.Event{}
	for i := 0; i < 100; i++ {
		if !s.Sample(event) {
			t.Fatal("expected all events to pass at rate 1.0")
		}
	}
}

func TestSample_RateNearZero_MostlyDropped(t *testing.T) {
	s, _ := New(Config{Rate: 0.01})
	event := &wal.Event{}
	passed := 0
	const trials = 10000
	for i := 0; i < trials; i++ {
		if s.Sample(event) {
			passed++
		}
	}
	// With rate 0.01 and 10000 trials, expect ~100 passes; allow generous bounds.
	if passed > 500 {
		t.Fatalf("expected few events to pass at rate 0.01, got %d/%d", passed, trials)
	}
}

func TestRate_ReturnsConfiguredValue(t *testing.T) {
	s, _ := New(Config{Rate: 0.75})
	if s.Rate() != 0.75 {
		t.Fatalf("expected rate 0.75, got %f", s.Rate())
	}
}
