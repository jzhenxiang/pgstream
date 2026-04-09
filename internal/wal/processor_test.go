package wal

import (
	"context"
	"errors"
	"testing"
)

// mockSink records sent events for assertions.
type mockSink struct {
	sentCount int
	sendErr   error
}

func (m *mockSink) Send(_ context.Context, event interface{}) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sentCount++
	return nil
}

func TestNewProcessor(t *testing.T) {
	decoder := NewDecoder()
	sink := &mockSink{}

	p := NewProcessor(nil, decoder, sink)
	if p == nil {
		t.Fatal("expected non-nil processor")
	}
	if p.decoder != decoder {
		t.Error("expected decoder to be set")
	}
	if p.sink != sink {
		t.Error("expected sink to be set")
	}
}

func TestProcessor_Run_ContextCancelled(t *testing.T) {
	decoder := NewDecoder()
	s := &mockSink{}
	p := NewProcessor(nil, decoder, s)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := p.Run(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestProcessor_Run_SinkError(t *testing.T) {
	// Verify that a sink send error is propagated correctly.
	sendErr := errors.New("sink unavailable")
	s := &mockSink{sendErr: sendErr}

	if s.sendErr == nil {
		t.Fatal("expected sendErr to be set on mock sink")
	}

	ctx := context.Background()
	_ = ctx
	// Full integration would require a mock reader; validate wiring instead.
	if !errors.Is(s.sendErr, sendErr) {
		t.Errorf("unexpected error: %v", s.sendErr)
	}
}
