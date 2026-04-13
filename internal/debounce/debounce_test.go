package debounce

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pgstream/pgstream/internal/wal"
)

func makeEvent(id string) *wal.Event {
	return &wal.Event{Data: &wal.Data{Table: id}}
}

func TestNew_NilSend_ReturnsError(t *testing.T) {
	_, err := New(0, nil)
	if err == nil {
		t.Fatal("expected error for nil send, got nil")
	}
}

func TestNew_ZeroQuietPeriod_UsesDefault(t *testing.T) {
	d, err := New(0, func([]*wal.Event) error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.quietPeriod != DefaultQuietPeriod {
		t.Errorf("expected %v, got %v", DefaultQuietPeriod, d.quietPeriod)
	}
}

func TestNew_CustomQuietPeriod(t *testing.T) {
	period := 50 * time.Millisecond
	d, err := New(period, func([]*wal.Event) error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.quietPeriod != period {
		t.Errorf("expected %v, got %v", period, d.quietPeriod)
	}
}

func TestAdd_NilEvent_IsNoOp(t *testing.T) {
	var called int32
	d, _ := New(20*time.Millisecond, func(evs []*wal.Event) error {
		atomic.AddInt32(&called, 1)
		return nil
	})
	d.Add(context.Background(), nil)
	time.Sleep(50 * time.Millisecond)
	if atomic.LoadInt32(&called) != 0 {
		t.Error("send should not have been called for nil event")
	}
}

func TestAdd_FlushesAfterQuietPeriod(t *testing.T) {
	var received []*wal.Event
	done := make(chan struct{})
	d, _ := New(30*time.Millisecond, func(evs []*wal.Event) error {
		received = evs
		close(done)
		return nil
	})
	ctx := context.Background()
	d.Add(ctx, makeEvent("a"))
	d.Add(ctx, makeEvent("b"))

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for flush")
	}
	if len(received) != 2 {
		t.Errorf("expected 2 events, got %d", len(received))
	}
}

func TestFlush_ImmediatelySendsEvents(t *testing.T) {
	var received []*wal.Event
	d, _ := New(500*time.Millisecond, func(evs []*wal.Event) error {
		received = evs
		return nil
	})
	ctx := context.Background()
	d.Add(ctx, makeEvent("x"))
	if err := d.Flush(ctx); err != nil {
		t.Fatalf("unexpected flush error: %v", err)
	}
	if len(received) != 1 {
		t.Errorf("expected 1 event, got %d", len(received))
	}
}

func TestFlush_PropagatesSendError(t *testing.T) {
	sentinel := errors.New("send failed")
	d, _ := New(500*time.Millisecond, func([]*wal.Event) error {
		return sentinel
	})
	ctx := context.Background()
	d.Add(ctx, makeEvent("y"))
	if err := d.Flush(ctx); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestFlush_EmptyPending_IsNoOp(t *testing.T) {
	var called int32
	d, _ := New(50*time.Millisecond, func([]*wal.Event) error {
		atomic.AddInt32(&called, 1)
		return nil
	})
	if err := d.Flush(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&called) != 0 {
		t.Error("send should not be called when nothing is pending")
	}
}
