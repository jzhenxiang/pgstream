package backpressure_test

import (
	"context"
	"testing"
	"time"

	"github.com/your-org/pgstream/internal/backpressure"
)

func TestNew_DefaultConfig(t *testing.T) {
	ctrl := backpressure.New(backpressure.DefaultConfig())
	if ctrl == nil {
		t.Fatal("expected non-nil controller")
	}
	if ctrl.Pending() != 0 {
		t.Fatalf("expected 0 pending, got %d", ctrl.Pending())
	}
}

func TestNew_ZeroMaxPending_UsesDefault(t *testing.T) {
	ctrl := backpressure.New(backpressure.Config{})
	if ctrl == nil {
		t.Fatal("expected non-nil controller")
	}
}

func TestAcquireRelease(t *testing.T) {
	ctrl := backpressure.New(backpressure.Config{
		MaxPending:     4,
		AcquireTimeout: time.Second,
	})
	ctx := context.Background()

	for i := 0; i < 4; i++ {
		if err := ctrl.Acquire(ctx); err != nil {
			t.Fatalf("unexpected error on acquire %d: %v", i, err)
		}
	}
	if ctrl.Pending() != 4 {
		t.Fatalf("expected 4 pending, got %d", ctrl.Pending())
	}

	for i := 0; i < 4; i++ {
		ctrl.Release()
	}
	if ctrl.Pending() != 0 {
		t.Fatalf("expected 0 pending after release, got %d", ctrl.Pending())
	}
}

func TestAcquire_Timeout(t *testing.T) {
	ctrl := backpressure.New(backpressure.Config{
		MaxPending:     1,
		AcquireTimeout: 50 * time.Millisecond,
	})
	ctx := context.Background()

	if err := ctrl.Acquire(ctx); err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}

	err := ctrl.Acquire(ctx)
	if err == nil {
		t.Fatal("expected backpressure error, got nil")
	}
	if err != backpressure.ErrBackpressure {
		t.Fatalf("expected ErrBackpressure, got %v", err)
	}
}

func TestAcquire_ContextCancelled(t *testing.T) {
	ctrl := backpressure.New(backpressure.Config{
		MaxPending:     1,
		AcquireTimeout: 5 * time.Second,
	})

	if err := ctrl.Acquire(context.Background()); err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := ctrl.Acquire(ctx)
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
}

func TestRelease_ExtraRelease_NoOp(t *testing.T) {
	ctrl := backpressure.New(backpressure.DefaultConfig())
	// Should not panic or block.
	ctrl.Release()
	if ctrl.Pending() != 0 {
		t.Fatalf("expected 0 pending, got %d", ctrl.Pending())
	}
}
