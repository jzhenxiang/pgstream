package buffer

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/pgstream/pgstream/internal/wal"
)

func makeEvent() *wal.Event {
	return &wal.Event{Table: "users", Action: "INSERT"}
}

func TestNew_DefaultConfig(t *testing.T) {
	buf := New(Config{}, func(_ []*wal.Event) error { return nil })
	if buf.cfg.MaxSize != DefaultMaxSize {
		t.Errorf("expected MaxSize %d, got %d", DefaultMaxSize, buf.cfg.MaxSize)
	}
	if buf.cfg.FlushInterval != DefaultFlushInterval {
		t.Errorf("expected FlushInterval %v, got %v", DefaultFlushInterval, buf.cfg.FlushInterval)
	}
}

func TestAdd_FlushesOnMaxSize(t *testing.T) {
	flushed := 0
	buf := New(Config{MaxSize: 3}, func(events []*wal.Event) error {
		flushed += len(events)
		return nil
	})
	for i := 0; i < 3; i++ {
		if err := buf.Add(makeEvent()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if flushed != 3 {
		t.Errorf("expected 3 flushed, got %d", flushed)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty buffer after flush, got %d", buf.Len())
	}
}

func TestAdd_NoFlushBelowMaxSize(t *testing.T) {
	flushed := 0
	buf := New(Config{MaxSize: 10}, func(events []*wal.Event) error {
		flushed += len(events)
		return nil
	})
	_ = buf.Add(makeEvent())
	_ = buf.Add(makeEvent())
	if flushed != 0 {
		t.Errorf("expected no flush, got %d", flushed)
	}
	if buf.Len() != 2 {
		t.Errorf("expected 2 buffered, got %d", buf.Len())
	}
}

func TestAdd_PropagatesFlushError(t *testing.T) {
	expected := errors.New("sink unavailable")
	buf := New(Config{MaxSize: 1}, func(_ []*wal.Event) error {
		return expected
	})
	if err := buf.Add(makeEvent()); !errors.Is(err, expected) {
		t.Errorf("expected sink error, got %v", err)
	}
}

func TestRun_PeriodicFlush(t *testing.T) {
	var mu sync.Mutex
	flushed := 0
	buf := New(Config{MaxSize: 100, FlushInterval: 20 * time.Millisecond}, func(events []*wal.Event) error {
		mu.Lock()
		flushed += len(events)
		mu.Unlock()
		return nil
	})
	_ = buf.Add(makeEvent())
	_ = buf.Add(makeEvent())

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	_ = buf.Run(ctx)

	mu.Lock()
	defer mu.Unlock()
	if flushed < 2 {
		t.Errorf("expected at least 2 flushed via ticker, got %d", flushed)
	}
}

func TestRun_FlushesRemainingOnCancel(t *testing.T) {
	flushed := 0
	buf := New(Config{MaxSize: 100, FlushInterval: time.Hour}, func(events []*wal.Event) error {
		flushed += len(events)
		return nil
	})
	_ = buf.Add(makeEvent())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = buf.Run(ctx)

	if flushed != 1 {
		t.Errorf("expected 1 flushed on cancel, got %d", flushed)
	}
}
