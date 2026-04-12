package splitter

import (
	"testing"

	"github.com/pgstream/pgstream/internal/wal"
)

func makeEvents(n int) []*wal.Event {
	events := make([]*wal.Event, n)
	for i := range events {
		events[i] = &wal.Event{}
	}
	return events
}

func TestNew_NegativeChunkSize_ReturnsError(t *testing.T) {
	_, err := New(Config{ChunkSize: -1})
	if err == nil {
		t.Fatal("expected error for negative chunk size, got nil")
	}
}

func TestNew_ZeroChunkSize_UsesDefault(t *testing.T) {
	s, err := New(Config{ChunkSize: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ChunkSize() != defaultChunkSize {
		t.Errorf("expected default chunk size %d, got %d", defaultChunkSize, s.ChunkSize())
	}
}

func TestNew_CustomChunkSize(t *testing.T) {
	s, err := New(Config{ChunkSize: 25})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ChunkSize() != 25 {
		t.Errorf("expected chunk size 25, got %d", s.ChunkSize())
	}
}

func TestSplit_NilEvents_ReturnsEmpty(t *testing.T) {
	s, _ := New(Config{ChunkSize: 10})
	chunks := s.Split(nil)
	if len(chunks) != 0 {
		t.Errorf("expected empty chunks, got %d", len(chunks))
	}
}

func TestSplit_ExactMultiple(t *testing.T) {
	s, _ := New(Config{ChunkSize: 5})
	events := makeEvents(10)
	chunks := s.Split(events)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	for i, c := range chunks {
		if len(c) != 5 {
			t.Errorf("chunk %d: expected 5 events, got %d", i, len(c))
		}
	}
}

func TestSplit_Remainder(t *testing.T) {
	s, _ := New(Config{ChunkSize: 4})
	events := makeEvents(9)
	chunks := s.Split(events)
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}
	if len(chunks[2]) != 1 {
		t.Errorf("last chunk: expected 1 event, got %d", len(chunks[2]))
	}
}

func TestSplit_DoesNotMutateOriginal(t *testing.T) {
	s, _ := New(Config{ChunkSize: 3})
	events := makeEvents(6)
	orig := make([]*wal.Event, len(events))
	copy(orig, events)
	s.Split(events)
	for i := range events {
		if events[i] != orig[i] {
			t.Errorf("original slice mutated at index %d", i)
		}
	}
}
