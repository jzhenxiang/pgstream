package throttle

import (
	"context"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.MaxDelay == 0 {
		t.Fatal("expected non-zero MaxDelay")
	}
	if cfg.Step == 0 {
		t.Fatal("expected non-zero Step")
	}
}

func TestNew_InvalidMaxDelay(t *testing.T) {
	_, err := New(Config{MinDelay: time.Second, MaxDelay: time.Millisecond, Step: time.Millisecond})
	if err == nil {
		t.Fatal("expected error when MaxDelay < MinDelay")
	}
}

func TestNew_Valid(t *testing.T) {
	th, err := New(DefaultConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if th == nil {
		t.Fatal("expected non-nil Throttle")
	}
}

func TestNew_ZeroValues_UsesDefaults(t *testing.T) {
	th, err := New(Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if th.cfg.MaxDelay == 0 {
		t.Fatal("expected MaxDelay to be set from defaults")
	}
	if th.cfg.Step == 0 {
		t.Fatal("expected Step to be set from defaults")
	}
}

func TestIncrease_CapsAtMax(t *testing.T) {
	cfg := Config{MinDelay: 0, MaxDelay: 200 * time.Millisecond, Step: 100 * time.Millisecond}
	th, _ := New(cfg)
	th.Increase()
	th.Increase()
	th.Increase() // would exceed max
	if th.Current() != cfg.MaxDelay {
		t.Fatalf("expected %v, got %v", cfg.MaxDelay, th.Current())
	}
}

func TestDecrease_FloorAtMin(t *testing.T) {
	cfg := Config{MinDelay: 50 * time.Millisecond, MaxDelay: 500 * time.Millisecond, Step: 100 * time.Millisecond}
	th, _ := New(cfg)
	th.Decrease() // already at min
	if th.Current() != cfg.MinDelay {
		t.Fatalf("expected %v, got %v", cfg.MinDelay, th.Current())
	}
}

func TestWait_ZeroDelay_ReturnsImmediately(t *testing.T) {
	th, _ := New(Config{MaxDelay: time.Second, Step: time.Millisecond})
	ctx := context.Background()
	start := time.Now()
	if err := th.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if time.Since(start) > 20*time.Millisecond {
		t.Fatal("Wait took too long with zero delay")
	}
}

func TestWait_ContextCancelled(t *testing.T) {
	cfg := Config{MinDelay: 0, MaxDelay: 5 * time.Second, Step: 5 * time.Second}
	th, _ := New(cfg)
	th.Increase() // set delay to 5s

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := th.Wait(ctx); err == nil {
		t.Fatal("expected context cancellation error")
	}
}

func TestIncreaseDecrease_RoundTrip(t *testing.T) {
	cfg := Config{MinDelay: 0, MaxDelay: 300 * time.Millisecond, Step: 100 * time.Millisecond}
	th, _ := New(cfg)
	th.Increase()
	th.Increase()
	if th.Current() != 200*time.Millisecond {
		t.Fatalf("expected 200ms, got %v", th.Current())
	}
	th.Decrease()
	if th.Current() != 100*time.Millisecond {
		t.Fatalf("expected 100ms, got %v", th.Current())
	}
}
