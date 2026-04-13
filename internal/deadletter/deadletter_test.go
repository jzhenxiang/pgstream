package deadletter_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pgstream/pgstream/internal/deadletter"
	"github.com/pgstream/pgstream/internal/wal"
)

func sampleEvent() *wal.Event {
	return &wal.Event{LSN: "0/1234"}
}

func TestNew_DefaultCapacity(t *testing.T) {
	q := deadletter.New(0)
	if q == nil {
		t.Fatal("expected non-nil queue")
	}
}

func TestNew_CustomCapacity(t *testing.T) {
	q := deadletter.New(10)
	if q.Len() != 0 {
		t.Fatalf("expected empty queue, got %d", q.Len())
	}
}

func TestPush_IncreasesLen(t *testing.T) {
	q := deadletter.New(5)
	q.Push(context.Background(), sampleEvent(), errors.New("boom"), 1)
	if q.Len() != 1 {
		t.Fatalf("expected len 1, got %d", q.Len())
	}
}

func TestPush_EvictsOldestWhenFull(t *testing.T) {
	const cap = 3
	q := deadletter.New(cap)
	ctx := context.Background()

	for i := 0; i < cap+1; i++ {
		e := &wal.Event{LSN: string(rune('A' + i))}
		q.Push(ctx, e, errors.New("err"), 1)
	}

	if q.Len() != cap {
		t.Fatalf("expected len %d after eviction, got %d", cap, q.Len())
	}
}

func TestDrain_ClearsQueue(t *testing.T) {
	q := deadletter.New(10)
	ctx := context.Background()
	q.Push(ctx, sampleEvent(), errors.New("e"), 1)
	q.Push(ctx, sampleEvent(), errors.New("e"), 2)

	entries := q.Drain()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if q.Len() != 0 {
		t.Fatalf("expected empty queue after drain, got %d", q.Len())
	}
}

func TestPeek_DoesNotClear(t *testing.T) {
	q := deadletter.New(10)
	q.Push(context.Background(), sampleEvent(), errors.New("e"), 1)

	_ = q.Peek()
	if q.Len() != 1 {
		t.Fatalf("expected queue to remain intact after Peek, got len %d", q.Len())
	}
}

func TestEntry_FieldsPopulated(t *testing.T) {
	q := deadletter.New(5)
	ev := sampleEvent()
	err := errors.New("sink unavailable")
	q.Push(context.Background(), ev, err, 3)

	entries := q.Drain()
	if len(entries) != 1 {
		t.Fatal("expected one entry")
	}
	e := entries[0]
	if e.Event != ev {
		t.Error("event pointer mismatch")
	}
	if e.Err != err {
		t.Error("error mismatch")
	}
	if e.Attempts != 3 {
		t.Errorf("expected attempts 3, got %d", e.Attempts)
	}
	if e.FailedAt.IsZero() {
		t.Error("expected FailedAt to be set")
	}
}
