package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/your-org/pgstream/internal/retry"
)

func TestDo_SuccessOnFirstAttempt(t *testing.T) {
	calls := 0
	err := retry.Do(context.Background(), retry.DefaultConfig(), func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_RetriesAndSucceeds(t *testing.T) {
	calls := 0
	sentinel := errors.New("transient")
	cfg := retry.Config{MaxAttempts: 3, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond, Multiplier: 2}

	err := retry.Do(context.Background(), cfg, func() error {
		calls++
		if calls < 3 {
			return sentinel
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error after retries, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ExhaustsAttempts(t *testing.T) {
	sentinel := errors.New("permanent")
	cfg := retry.Config{MaxAttempts: 3, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond, Multiplier: 2}
	calls := 0

	err := retry.Do(context.Background(), cfg, func() error {
		calls++
		return sentinel
	})
	if !errors.Is(err, retry.ErrMaxAttemptsReached) {
		t.Fatalf("expected ErrMaxAttemptsReached, got %v", err)
	}
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected wrapped sentinel error, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calls := 0
	err := retry.Do(ctx, retry.DefaultConfig(), func() error {
		calls++
		return nil
	})
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
	if calls != 0 {
		t.Fatalf("expected 0 calls after pre-cancelled context, got %d", calls)
	}
}

func TestDo_ZeroMaxAttempts_RunsOnce(t *testing.T) {
	calls := 0
	cfg := retry.Config{MaxAttempts: 0, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond, Multiplier: 2}
	err := retry.Do(context.Background(), cfg, func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}
