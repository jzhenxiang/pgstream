package tee_test

import (
	"context"
	"errors"
	"testing"

	"pgstream/internal/tee"
	"pgstream/internal/wal"
)

type mockSink struct {
	err     error
	called  bool
	payload *wal.Event
}

func (m *mockSink) Send(_ context.Context, e *wal.Event) error {
	m.called = true
	m.payload = e
	return m.err
}

func TestNew_NoSinks_ReturnsError(t *testing.T) {
	_, err := tee.New()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNew_NilSink_ReturnsError(t *testing.T) {
	_, err := tee.New(&mockSink{}, nil)
	if err == nil {
		t.Fatal("expected error for nil sink")
	}
}

func TestNew_Valid(t *testing.T) {
	t1, err := tee.New(&mockSink{}, &mockSink{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if t1.Len() != 2 {
		t.Fatalf("expected 2 sinks, got %d", t1.Len())
	}
}

func TestSend_AllSinksReceiveEvent(t *testing.T) {
	a, b := &mockSink{}, &mockSink{}
	tee, _ := tee.New(a, b)
	event := &wal.Event{}
	if err := tee.Send(context.Background(), event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !a.called || !b.called {
		t.Fatal("expected both sinks to be called")
	}
}

func TestSend_OneSinkFails_OtherStillCalled(t *testing.T) {
	boom := errors.New("sink exploded")
	a := &mockSink{err: boom}
	b := &mockSink{}
	tee, _ := tee.New(a, b)
	err := tee.Send(context.Background(), &wal.Event{})
	if err == nil {
		t.Fatal("expected combined error")
	}
	if !b.called {
		t.Fatal("second sink should still be called even after first fails")
	}
	if !errors.Is(err, boom) {
		t.Fatalf("expected wrapped boom error, got: %v", err)
	}
}

func TestSend_AllSinksFail_ReturnsJoinedError(t *testing.T) {
	e1 := errors.New("err1")
	e2 := errors.New("err2")
	a := &mockSink{err: e1}
	b := &mockSink{err: e2}
	tee, _ := tee.New(a, b)
	err := tee.Send(context.Background(), &wal.Event{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, e1) || !errors.Is(err, e2) {
		t.Fatalf("expected both errors joined, got: %v", err)
	}
}
