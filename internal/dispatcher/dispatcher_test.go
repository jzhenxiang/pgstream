package dispatcher

import (
	"context"
	"errors"
	"testing"

	"github.com/pgstream/pgstream/internal/wal"
)

// mockSink records the last event it received.
type mockSink struct {
	received []*wal.Event
	err      error
}

func (m *mockSink) Send(_ context.Context, e *wal.Event) error {
	if m.err != nil {
		return m.err
	}
	m.received = append(m.received, e)
	return nil
}

func TestNew_NoRoutesNoDefaults_ReturnsError(t *testing.T) {
	_, err := New(nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNew_EmptyTableName_ReturnsError(t *testing.T) {
	s := &mockSink{}
	_, err := New([]Route{{Table: "", Sinks: []interface{}{s}}}, nil)
	if err == nil {
		t.Fatal("expected error for empty table name")
	}
}

func TestNew_RouteWithNoSinks_ReturnsError(t *testing.T) {
	_, err := New([]Route{{Table: "public.orders", Sinks: nil}}, nil)
	if err == nil {
		t.Fatal("expected error for route with no sinks")
	}
}

func TestDispatch_MatchingRoute_SendsToRouteSink(t *testing.T) {
	routed := &mockSink{}
	defaultSink := &mockSink{}

	d, err := New([]Route{{Table: "public.orders", Sinks: []interface{}{routed}}}, []interface{}{defaultSink})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	event := &wal.Event{Table: "public.orders"}
	if err := d.Dispatch(context.Background(), event); err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	if len(routed.received) != 1 {
		t.Errorf("routed sink: expected 1 event, got %d", len(routed.received))
	}
	if len(defaultSink.received) != 0 {
		t.Errorf("default sink should not receive event")
	}
}

func TestDispatch_NoMatchingRoute_SendsToDefault(t *testing.T) {
	routed := &mockSink{}
	defaultSink := &mockSink{}

	d, err := New([]Route{{Table: "public.orders", Sinks: []interface{}{routed}}}, []interface{}{defaultSink})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	event := &wal.Event{Table: "public.users"}
	if err := d.Dispatch(context.Background(), event); err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	if len(defaultSink.received) != 1 {
		t.Errorf("default sink: expected 1 event, got %d", len(defaultSink.received))
	}
	if len(routed.received) != 0 {
		t.Errorf("routed sink should not receive event")
	}
}

func TestDispatch_SinkError_ReturnsError(t *testing.T) {
	s := &mockSink{err: errors.New("sink failure")}
	d, err := New([]Route{{Table: "public.orders", Sinks: []interface{}{s}}}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	event := &wal.Event{Table: "public.orders"}
	if err := d.Dispatch(context.Background(), event); err == nil {
		t.Fatal("expected error from sink, got nil")
	}
}

func TestDispatch_NilEvent_UsesDefaults(t *testing.T) {
	defaultSink := &mockSink{}
	d, err := New(nil, []interface{}{defaultSink})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := d.Dispatch(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(defaultSink.received) != 1 {
		t.Errorf("default sink: expected 1 event, got %d", len(defaultSink.received))
	}
}
