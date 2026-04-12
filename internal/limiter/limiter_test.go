package limiter

import (
	"context"
	"testing"
	"time"
)

func TestNew_NegativeRate_ReturnsError(t *testing.T) {
	_, err := New(Config{MaxEventsPerSecond: -1})
	if err == nil {
		t.Fatal("expected error for negative rate, got nil")
	}
}

func TestNew_ZeroRate_Disabled(t *testing.T) {
	lim, err := New(Config{MaxEventsPerSecond: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should never be limited.
	for i := 0; i < 1000; i++ {
		if err := lim.Allow(context.Background()); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	}
}

func TestNew_DefaultBurstEqualsRate(t *testing.T) {
	lim, err := New(Config{MaxEventsPerSecond: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lim.cfg.BurstSize != 5 {
		t.Fatalf("expected burst 5, got %d", lim.cfg.BurstSize)
	}
}

func TestAllow_ConsumesTokens(t *testing.T) {
	lim, err := New(Config{MaxEventsPerSecond: 3, BurstSize: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if err := lim.Allow(ctx); err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
	}
	// Bucket should now be empty.
	if err := lim.Allow(ctx); err == nil {
		t.Fatal("expected ErrLimitExceeded after burst exhausted")
	}
}

func TestAllow_RefillsOverTime(t *testing.T) {
	lim, err := New(Config{MaxEventsPerSecond: 100, BurstSize: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := context.Background()
	// Drain the single token.
	_ = lim.Allow(ctx)
	// Wait long enough for at least one token to refill.
	time.Sleep(20 * time.Millisecond)
	if err := lim.Allow(ctx); err != nil {
		t.Fatalf("expected token to refill, got: %v", err)
	}
}

func TestReset_RestoresBurst(t *testing.T) {
	lim, err := New(Config{MaxEventsPerSecond: 2, BurstSize: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := context.Background()
	_ = lim.Allow(ctx)
	_ = lim.Allow(ctx)
	// Exhausted – reset should restore.
	lim.Reset()
	if err := lim.Allow(ctx); err != nil {
		t.Fatalf("expected allow after reset, got: %v", err)
	}
}
