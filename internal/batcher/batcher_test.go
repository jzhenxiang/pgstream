package batcher_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pgstream/pgstream/internal/batcher"
	"github.com/pgstream/pgstream/internal/wal"
)

func makeEvent(id string) *wal.Event {
	return &wal.Event{Table: id}
}

func TestNew_NilSend_ReturnsError(t *testing.T) {
	_, err := batcher.New(batcher.Config{}, nil)
	if !errors.Is(err, batcher.ErrNoSink) {
		t.Fatalf("expected ErrNoSink, got %v", err)
	}
}

func TestNew_Valid(t *testing.T) {
	b, err := batcher.New(batcher.Config{MaxSize: 10, FlushInterval: time.Second}, func(_ context.Context, _ []*wal.Event) error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b == nil {
		t.Fatal("expected non-nil batcher")
	}
}

func TestRun_FlushesOnMaxSize(t *testing.T) {
	var flushed atomic.Int32
	ch := make(chan *wal.Event, 20)

	b, _ := batcher.New(batcher.Config{MaxSize: 5, FlushInterval: 10 * time.Second}, func(_ context.Context, evs []*wal.Event) error {
		flushed.Add(int32(len(evs)))
		return nil
	})

	for i := 0; i < 10; i++ {
		ch <- makeEvent("t")
	}
	close(ch)

	ctx := context.Background()
	if err := b.Run(ctx, ch); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := flushed.Load(); got != 10 {
		t.Fatalf("expected 10 flushed events, got %d", got)
	}
}

func TestRun_FlushesOnTicker(t *testing.T) {
	var flushed atomic.Int32
	ch := make(chan *wal.Event, 10)

	b, _ := batcher.New(batcher.Config{MaxSize: 100, FlushInterval: 50 * time.Millisecond}, func(_ context.Context, evs []*wal.Event) error {
		flushed.Add(int32(len(evs)))
		return nil
	})

	ch <- makeEvent("t")
	ch <- makeEvent("t")

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	b.Run(ctx, ch) //nolint:errcheck

	if got := flushed.Load(); got < 2 {
		t.Fatalf("expected at least 2 flushed events, got %d", got)
	}
}

func TestRun_ContextCancelled_FlushesRemainder(t *testing.T) {
	var flushed atomic.Int32
	ch := make(chan *wal.Event, 5)

	b, _ := batcher.New(batcher.Config{MaxSize: 100, FlushInterval: 10 * time.Second}, func(_ context.Context, evs []*wal.Event) error {
		flushed.Add(int32(len(evs)))
		return nil
	})

	ch <- makeEvent("t")
	ch <- makeEvent("t")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	b.Run(ctx, ch) //nolint:errcheck

	if got := flushed.Load(); got != 2 {
		t.Fatalf("expected 2 flushed events on cancel, got %d", got)
	}
}

func TestRun_SendError_Propagates(t *testing.T) {
	sentinel := errors.New("send failed")
	ch := make(chan *wal.Event, 5)

	b, _ := batcher.New(batcher.Config{MaxSize: 2, FlushInterval: 10 * time.Second}, func(_ context.Context, _ []*wal.Event) error {
		return sentinel
	})

	ch <- makeEvent("t")
	ch <- makeEvent("t")

	err := b.Run(context.Background(), ch)
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
