package semaphore

import (
	"testing"
)

func TestMax_ReturnsConfiguredMax(t *testing.T) {
	sem, _ := New(7)
	if sem.Max() != 7 {
		t.Errorf("expected 7, got %d", sem.Max())
	}
}

func TestCurrent_InitiallyZero(t *testing.T) {
	sem, _ := New(3)
	if sem.Current() != 0 {
		t.Errorf("expected 0, got %d", sem.Current())
	}
}

func TestRelease_BelowZero_IsNoOp(t *testing.T) {
	sem, _ := New(2)
	// Release without acquire should not panic or go negative.
	sem.Release()
	if sem.Current() != 0 {
		t.Errorf("expected 0, got %d", sem.Current())
	}
}

func TestAcquireUpToMax(t *testing.T) {
	const n = 4
	sem, _ := New(n)
	for i := 0; i < n; i++ {
		if err := sem.TryAcquire(); err != nil {
			t.Fatalf("slot %d: unexpected error: %v", i, err)
		}
	}
	if err := sem.TryAcquire(); err != ErrAcquire {
		t.Errorf("expected ErrAcquire when full, got %v", err)
	}
	if sem.Current() != n {
		t.Errorf("expected current=%d, got %d", n, sem.Current())
	}
}
