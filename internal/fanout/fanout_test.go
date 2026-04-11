package fanout

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/pgstream/pgstream/internal/wal"
)

// mockSink is a test double for sink.Sink.
type mockSink struct {
	sendFn func(ctx context.Context, event *wal.Event) error
	calls  atomic.Int32
}

func (m *mockSink) Send(ctx context.Context, event *wal.Event) error {
	m.calls.Add(1)
	if m.sendFn != nil {
		return m.sendFn(ctx, event)
	}
	return nil
}

func TestNew_NoSinks_ReturnsError(t *testing.T) {
	_, err := New()
	if err == nil {
		t.Fatal("expected error for empty sinks, got nil")
	}
}

func TestNew_NilSink_ReturnsError(t *testing.T) {
	_, err := New(nil)
	if err == nil {
		t.Fatal("expected error for nil sink, got nil")
	}
}

func TestNew_Valid(t *testing.T) {
	s := &mockSink{}
	f, err := New(s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Len() != 1 {
		t.Fatalf("expected 1 sink, got %d", f.Len())
	}
}

func TestSend_AllSinksReceiveEvent(t *testing.T) {
	s1 := &mockSink{}
	s2 := &mockSink{}
	f, _ := New(s1, s2)

	event := &wal.Event{}
	if err := f.Send(context.Background(), event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s1.calls.Load() != 1 {
		t.Errorf("s1: expected 1 call, got %d", s1.calls.Load())
	}
	if s2.calls.Load() != 1 {
		t.Errorf("s2: expected 1 call, got %d", s2.calls.Load())
	}
}

func TestSend_OneSinkFails_ReturnsError(t *testing.T) {
	s1 := &mockSink{}
	s2 := &mockSink{sendFn: func(_ context.Context, _ *wal.Event) error {
		return errors.New("sink2 down")
	}}
	f, _ := New(s1, s2)

	err := f.Send(context.Background(), &wal.Event{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "sink2 down") {
		t.Errorf("error missing sink message: %v", err)
	}
	// s1 must still have been called
	if s1.calls.Load() != 1 {
		t.Errorf("s1 should still be called, got %d calls", s1.calls.Load())
	}
}

func TestSend_AllSinksFail_CombinesErrors(t *testing.T) {
	s1 := &mockSink{sendFn: func(_ context.Context, _ *wal.Event) error { return errors.New("err1") }}
	s2 := &mockSink{sendFn: func(_ context.Context, _ *wal.Event) error { return errors.New("err2") }}
	f, _ := New(s1, s2)

	err := f.Send(context.Background(), &wal.Event{})
	if err == nil {
		t.Fatal("expected combined error, got nil")
	}
	if !strings.Contains(err.Error(), "2 sink(s) failed") {
		t.Errorf("unexpected error format: %v", err)
	}
}
