package aggregator

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pgstream/pgstream/internal/wal"
)

func makeEvent(table string) *wal.Event {
	return &wal.Event{Table: table}
}

func TestNew_NilFlush_ReturnsError(t *testing.T) {
	_, err := New(Config{}, nil)
	if err == nil {
		t.Fatal("expected error for nil flush func")
	}
}

func TestNew_NegativeWindowSize_ReturnsError(t *testing.T) {
	_, err := New(Config{WindowSize: -1}, func(_ context.Context, _ string, _ []*wal.Event) error { return nil })
	if err == nil {
		t.Fatal("expected error for negative window size")
	}
}

func TestNew_DefaultsApplied(t *testing.T) {
	agg, err := New(Config{}, func(_ context.Context, _ string, _ []*wal.Event) error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if agg.windowSize != defaultWindowSize {
		t.Errorf("expected default window size %d, got %d", defaultWindowSize, agg.windowSize)
	}
	if agg.flushInterval != defaultFlushInterval {
		t.Errorf("expected default flush interval %v, got %v", defaultFlushInterval, agg.flushInterval)
	}
}

func TestAdd_NilEvent_IsNoOp(t *testing.T) {
	var called bool
	agg, _ := New(Config{WindowSize: 1}, func(_ context.Context, _ string, _ []*wal.Event) error {
		called = true
		return nil
	})
	if err := agg.Add(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("flush should not be called for nil event")
	}
}

func TestAdd_FlushesOnWindowSize(t *testing.T) {
	var flushed atomic.Int32
	agg, _ := New(Config{WindowSize: 3}, func(_ context.Context, _ string, evs []*wal.Event) error {
		flushed.Add(int32(len(evs)))
		return nil
	})
	for i := 0; i < 3; i++ {
		if err := agg.Add(context.Background(), makeEvent("orders")); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if flushed.Load() != 3 {
		t.Errorf("expected 3 flushed events, got %d", flushed.Load())
	}
}

func TestRun_FlushesOnTicker(t *testing.T) {
	var flushed atomic.Int32
	agg, _ := New(Config{WindowSize: 100, FlushInterval: 20 * time.Millisecond},
		func(_ context.Context, _ string, evs []*wal.Event) error {
			flushed.Add(int32(len(evs)))
			return nil
		})

	_ = agg.Add(context.Background(), makeEvent("users"))

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	_ = agg.Run(ctx)

	if flushed.Load() == 0 {
		t.Error("expected at least one ticker-driven flush")
	}
}

func TestRun_StopsOnContextCancel(t *testing.T) {
	agg, _ := New(Config{FlushInterval: 10 * time.Millisecond},
		func(_ context.Context, _ string, _ []*wal.Event) error { return nil })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := agg.Run(ctx)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}
