package slicer

import (
	"errors"
	"sync/atomic"
	"testing"

	"github.com/your-org/pgstream/internal/wal"
)

func makeEvent(id string) *wal.Event {
	return &wal.Event{ID: id}
}

func TestNew_NilFlush_ReturnsError(t *testing.T) {
	_, err := New(Config{}, nil)
	if err == nil {
		t.Fatal("expected error for nil flush, got nil")
	}
}

func TestNew_DefaultsApplied(t *testing.T) {
	s, err := New(Config{}, func(_ []*wal.Event) error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.cfg.MaxSize != DefaultMaxSize {
		t.Errorf("expected MaxSize %d, got %d", DefaultMaxSize, s.cfg.MaxSize)
	}
	if s.cfg.Interval != DefaultInterval {
		t.Errorf("expected Interval %v, got %v", DefaultInterval, s.cfg.Interval)
	}
}

func TestAdd_NilEvent_IsNoOp(t *testing.T) {
	var called int32
	s, _ := New(Config{MaxSize: 1}, func(_ []*wal.Event) error {
		atomic.AddInt32(&called, 1)
		return nil
	})
	if err := s.Add(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&called) != 0 {
		t.Error("flush should not be called for nil event")
	}
}

func TestAdd_FlushesOnMaxSize(t *testing.T) {
	var flushed [][]*wal.Event
	s, _ := New(Config{MaxSize: 3}, func(slice []*wal.Event) error {
		flushed = append(flushed, slice)
		return nil
	})
	for i := 0; i < 3; i++ {
		if err := s.Add(makeEvent("e")); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if len(flushed) != 1 {
		t.Fatalf("expected 1 flush, got %d", len(flushed))
	}
	if len(flushed[0]) != 3 {
		t.Errorf("expected slice of 3, got %d", len(flushed[0]))
	}
}

func TestAdd_NoFlushBelowMaxSize(t *testing.T) {
	var called int32
	s, _ := New(Config{MaxSize: 10}, func(_ []*wal.Event) error {
		atomic.AddInt32(&called, 1)
		return nil
	})
	for i := 0; i < 5; i++ {
		_ = s.Add(makeEvent("e"))
	}
	if atomic.LoadInt32(&called) != 0 {
		t.Error("flush should not be called below max size")
	}
}

func TestAdd_FlushErrorPropagates(t *testing.T) {
	expected := errors.New("add flush error")
	s, _ := New(Config{MaxSize: 2}, func(_ []*wal.Event) error { return expected })
	_ = s.Add(makeEvent("a"))
	err := s.Add(makeEvent("b"))
	if !errors.Is(err, expected) {
		t.Errorf("expected %v, got %v", expected, err)
	}
}

func TestFlush_EmptyBuffer_IsNoOp(t *testing.T) {
	var called int32
	s, _ := New(Config{}, func(_ []*wal.Event) error {
		atomic.AddInt32(&called, 1)
		return nil
	})
	if err := s.Flush(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&called) != 0 {
		t.Error("flush should not be called on empty buffer")
	}
}

func TestFlush_PropagatesError(t *testing.T) {
	expected := errors.New("flush error")
	s, _ := New(Config{}, func(_ []*wal.Event) error { return expected })
	_ = s.Add(makeEvent("e"))
	if err := s.Flush(); !errors.Is(err, expected) {
		t.Errorf("expected %v, got %v", expected, err)
	}
}

func TestFlush_ClearsBuffer(t *testing.T) {
	var count int32
	s, _ := New(Config{}, func(slice []*wal.Event) error {
		atomic.AddInt32(&count, int32(len(slice)))
		return nil
	})
	_ = s.Add(makeEvent("a"))
	_ = s.Add(makeEvent("b"))
	_ = s.Flush()
	_ = s.Flush() // second flush should be no-op
	if atomic.LoadInt32(&count) != 2 {
		t.Errorf("expected 2 events flushed total, got %d", count)
	}
}
