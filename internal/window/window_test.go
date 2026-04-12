package window

import (
	"testing"
	"time"
)

func TestNew_InvalidSize(t *testing.T) {
	_, err := New(0, time.Second)
	if err != ErrInvalidSize {
		t.Fatalf("expected ErrInvalidSize, got %v", err)
	}
}

func TestNew_InvalidGranule_Zero(t *testing.T) {
	_, err := New(time.Minute, 0)
	if err != ErrInvalidGranule {
		t.Fatalf("expected ErrInvalidGranule, got %v", err)
	}
}

func TestNew_InvalidGranule_ExceedsSize(t *testing.T) {
	_, err := New(time.Second, 2*time.Second)
	if err != ErrInvalidGranule {
		t.Fatalf("expected ErrInvalidGranule, got %v", err)
	}
}

func TestNew_Valid(t *testing.T) {
	w, err := New(time.Minute, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w == nil {
		t.Fatal("expected non-nil window")
	}
}

func TestCount_InitiallyZero(t *testing.T) {
	w, _ := New(time.Minute, time.Second)
	if c := w.Count(); c != 0 {
		t.Fatalf("expected 0, got %d", c)
	}
}

func TestAdd_And_Count(t *testing.T) {
	w, _ := New(time.Minute, time.Second)
	w.Add(3)
	w.Add(7)
	if c := w.Count(); c != 10 {
		t.Fatalf("expected 10, got %d", c)
	}
}

func TestReset_ClearsCount(t *testing.T) {
	w, _ := New(time.Minute, time.Second)
	w.Add(5)
	w.Reset()
	if c := w.Count(); c != 0 {
		t.Fatalf("expected 0 after reset, got %d", c)
	}
}

func TestEvict_RemovesOldBuckets(t *testing.T) {
	// Use a very small window so we can inject old entries directly.
	w, _ := New(100*time.Millisecond, 10*time.Millisecond)
	// Inject a stale bucket manually.
	w.mu.Lock()
	w.buckets = append(w.buckets, entry{
		at:    time.Now().Add(-200 * time.Millisecond).Truncate(10 * time.Millisecond),
		count: 99,
	})
	w.mu.Unlock()

	w.Add(1)
	if c := w.Count(); c != 1 {
		t.Fatalf("expected stale bucket evicted, count=1, got %d", c)
	}
}

func TestAdd_SameBucket_Accumulates(t *testing.T) {
	w, _ := New(time.Minute, time.Second)
	// Both adds happen within the same truncated second.
	w.Add(4)
	w.Add(6)
	if len(w.buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(w.buckets))
	}
	if w.buckets[0].count != 10 {
		t.Fatalf("expected bucket count 10, got %d", w.buckets[0].count)
	}
}
