package semaphore

import (
	"context"
	"testing"
	"time"
)

func TestNew_InvalidMax(t *testing.T) {
	_, err := New(0)
	if err == nil {
		t.Fatal("expected error for max=0")
	}
	_, err = New(-1)
	if err == nil {
		t.Fatal("expected error for max=-1")
	}
}

func TestNew_Valid(t *testing.T) {
	sem, err := New(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sem.Max() != 5 {
		t.Errorf("expected max=5, got %d", sem.Max())
	}
	if sem.Current() != 0 {
		t.Errorf("expected current=0, got %d", sem.Current())
	}
}

func TestAcquireRelease(t *testing.T) {
	sem, _ := New(2)
	ctx := context.Background()

	if err := sem.Acquire(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sem.Current() != 1 {
		t.Errorf("expected current=1, got %d", sem.Current())
	}

	if err := sem.Acquire(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sem.Current() != 2 {
		t.Errorf("expected current=2, got %d", sem.Current())
	}

	sem.Release()
	if sem.Current() != 1 {
		t.Errorf("expected current=1 after release, got %d", sem.Current())
	}
}

func TestAcquire_ContextCancelled(t *testing.T) {
	sem, _ := New(1)
	ctx := context.Background()

	// Fill the semaphore.
	_ = sem.Acquire(ctx)

	ctx2, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := sem.Acquire(ctx2)
	if err == nil {
		t.Fatal("expected context error")
	}
}

func TestTryAcquire_Success(t *testing.T) {
	sem, _ := New(2)
	if err := sem.TryAcquire(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sem.Current() != 1 {
		t.Errorf("expected current=1, got %d", sem.Current())
	}
}

func TestTryAcquire_Full(t *testing.T) {
	sem, _ := New(1)
	_ = sem.TryAcquire()

	err := sem.TryAcquire()
	if err != ErrAcquire {
		t.Errorf("expected ErrAcquire, got %v", err)
	}
}
