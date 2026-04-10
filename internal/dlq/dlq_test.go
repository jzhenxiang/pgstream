package dlq_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/pgstream/pgstream/internal/dlq"
	"github.com/pgstream/pgstream/internal/wal"
)

func TestNew_ReturnsEmptyQueue(t *testing.T) {
	q := dlq.New(dlq.Config{})
	if q == nil {
		t.Fatal("expected non-nil DLQ")
	}
	if q.Size() != 0 {
		t.Fatalf("expected size 0, got %d", q.Size())
	}
}

func TestPush_InMemory(t *testing.T) {
	q := dlq.New(dlq.Config{})
	event := &wal.Event{Table: "users"}

	if err := q.Push(event, errors.New("sink error"), 3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if q.Size() != 1 {
		t.Fatalf("expected size 1, got %d", q.Size())
	}

	entries := q.Entries()
	if entries[0].Error != "sink error" {
		t.Errorf("expected error 'sink error', got %q", entries[0].Error)
	}
	if entries[0].Attempts != 3 {
		t.Errorf("expected attempts 3, got %d", entries[0].Attempts)
	}
	if entries[0].Event.Table != "users" {
		t.Errorf("expected table 'users', got %q", entries[0].Event.Table)
	}
}

func TestPush_MultiplePushes(t *testing.T) {
	q := dlq.New(dlq.Config{})
	for i := 0; i < 5; i++ {
		if err := q.Push(&wal.Event{}, errors.New("err"), 1); err != nil {
			t.Fatalf("push %d failed: %v", i, err)
		}
	}
	if q.Size() != 5 {
		t.Fatalf("expected size 5, got %d", q.Size())
	}
}

func TestPush_PersistsToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dlq.jsonl")

	q := dlq.New(dlq.Config{FilePath: path})
	if err := q.Push(&wal.Event{Table: "orders"}, errors.New("timeout"), 2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read dlq file: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty dlq file")
	}
}

func TestEntries_ReturnsCopy(t *testing.T) {
	q := dlq.New(dlq.Config{})
	_ = q.Push(&wal.Event{Table: "t1"}, errors.New("e"), 1)

	a := q.Entries()
	a[0].Error = "mutated"

	b := q.Entries()
	if b[0].Error == "mutated" {
		t.Error("Entries() should return a copy, not a reference")
	}
}
