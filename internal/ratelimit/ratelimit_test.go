package ratelimit_test

import (
	"context"
	"testing"
	"time"

	"github.com/your-org/pgstream/internal/ratelimit"
)

func TestNew_NegativeRate_ReturnsError(t *testing.T) {
	_, err := ratelimit.New(ratelimit.Config{EventsPerSecond: -1})
	if err == nil {
		t.Fatal("expected error for negative EventsPerSecond, got nil")
	}
}

func TestNew_ZeroRate_Disabled(t *testing.T) {
	l, err := ratelimit.New(ratelimit.Config{EventsPerSecond: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer l.Stop()

	ctx := context.Background()
	// Should return immediately without blocking.
	done := make(chan error, 1)
	go func() { done <- l.Wait(ctx) }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Wait blocked unexpectedly on disabled limiter")
	}
}

func TestWait_ConsumesToken(t *testing.T) {
	l, err := ratelimit.New(ratelimit.Config{EventsPerSecond: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer l.Stop()

	ctx := context.Background()
	if err := l.Wait(ctx); err != nil {
		t.Fatalf("Wait returned error: %v", err)
	}
}

func TestWait_ContextCancelled(t *testing.T) {
	// Rate of 1 per second; drain the initial token first.
	l, err := ratelimit.New(ratelimit.Config{EventsPerSecond: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer l.Stop()

	ctx := context.Background()
	// Drain the pre-filled token.
	_ = l.Wait(ctx)

	// Now cancel context before next token is available.
	ctxCancel, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err = l.Wait(ctxCancel)
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
}

func TestStop_StopsLimiter(t *testing.T) {
	l, err := ratelimit.New(ratelimit.Config{EventsPerSecond: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Drain the pre-filled token.
	_ = l.Wait(context.Background())

	l.Stop()

	// After stop, Wait should return an error quickly.
	done := make(chan error, 1)
	go func() { done <- l.Wait(context.Background()) }()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected error after Stop, got nil")
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Wait did not return after Stop")
	}
}
