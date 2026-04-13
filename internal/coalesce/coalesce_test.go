package coalesce

import (
	"sync"
	"testing"
	"time"

	"pgstream/internal/wal"
)

func makeEvent(table, rowID, value string) *wal.Event {
	return &wal.Event{Table: table, RowID: rowID, Data: map[string]any{"v": value}}
}

func TestNew_NilFlush_ReturnsError(t *testing.T) {
	_, err := New(0, 0, nil)
	if err == nil {
		t.Fatal("expected error for nil flush func")
	}
}

func TestNew_DefaultsApplied(t *testing.T) {
	c, err := New(0, 0, func(*wal.Event) error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.maxSize != defaultWindowSize {
		t.Errorf("maxSize: got %d, want %d", c.maxSize, defaultWindowSize)
	}
	if c.quietPeriod != defaultQuietPeriod {
		t.Errorf("quietPeriod: got %v, want %v", c.quietPeriod, defaultQuietPeriod)
	}
}

func TestAdd_NilEvent_IsNoOp(t *testing.T) {
	c, _ := New(10, 10*time.Millisecond, func(*wal.Event) error { return nil })
	if err := c.Add(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAdd_CoalescesUpdates(t *testing.T) {
	var mu sync.Mutex
	var flushed []*wal.Event

	c, _ := New(10, 20*time.Millisecond, func(ev *wal.Event) error {
		mu.Lock()
		flushed = append(flushed, ev)
		mu.Unlock()
		return nil
	})

	_ = c.Add(makeEvent("users", "1", "first"))
	_ = c.Add(makeEvent("users", "1", "second"))
	_ = c.Add(makeEvent("users", "1", "third"))

	time.Sleep(60 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(flushed) != 1 {
		t.Fatalf("expected 1 flushed event, got %d", len(flushed))
	}
	if flushed[0].Data["v"] != "third" {
		t.Errorf("expected last value 'third', got %v", flushed[0].Data["v"])
	}
}

func TestAdd_EvictsOldestWhenFull(t *testing.T) {
	var mu sync.Mutex
	var flushed []*wal.Event

	c, _ := New(2, 500*time.Millisecond, func(ev *wal.Event) error {
		mu.Lock()
		flushed = append(flushed, ev)
		mu.Unlock()
		return nil
	})

	_ = c.Add(makeEvent("t", "a", "A"))
	_ = c.Add(makeEvent("t", "b", "B"))
	// Adding a third distinct key forces eviction of "a".
	_ = c.Add(makeEvent("t", "c", "C"))

	mu.Lock()
	defer mu.Unlock()
	if len(flushed) != 1 {
		t.Fatalf("expected 1 evicted event, got %d", len(flushed))
	}
	if flushed[0].RowID != "a" {
		t.Errorf("expected evicted rowID 'a', got %s", flushed[0].RowID)
	}
}

func TestAdd_DifferentKeys_FlushedSeparately(t *testing.T) {
	var mu sync.Mutex
	var flushed []*wal.Event

	c, _ := New(10, 20*time.Millisecond, func(ev *wal.Event) error {
		mu.Lock()
		flushed = append(flushed, ev)
		mu.Unlock()
		return nil
	})

	_ = c.Add(makeEvent("orders", "10", "x"))
	_ = c.Add(makeEvent("orders", "20", "y"))

	time.Sleep(60 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(flushed) != 2 {
		t.Fatalf("expected 2 flushed events, got %d", len(flushed))
	}
}
