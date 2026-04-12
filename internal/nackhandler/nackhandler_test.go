package nackhandler

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/pgstream/internal/wal"
)

// stubSink records the last event it received.
type stubSink struct {
	received []*wal.Event
	err      error
}

func (s *stubSink) Send(_ context.Context, e *wal.Event) error {
	s.received = append(s.received, e)
	return s.err
}

func sampleEvent() *wal.Event {
	return &wal.Event{Schema: "public", Table: "orders", LSN: "0/1A2B3C"}
}

func TestNew_NilFallback_ReturnsError(t *testing.T) {
	_, err := New(Config{})
	if err == nil {
		t.Fatal("expected error for nil fallback, got nil")
	}
}

func TestNew_DefaultMaxAttempts(t *testing.T) {
	fb := &stubSink{}
	h, err := New(Config{Fallback: fb})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.maxAttempts != DefaultMaxAttempts {
		t.Fatalf("expected %d, got %d", DefaultMaxAttempts, h.maxAttempts)
	}
}

func TestHandle_NilEvent_ReturnsError(t *testing.T) {
	h, _ := New(Config{Fallback: &stubSink{}})
	if err := h.Handle(context.Background(), nil, errors.New("oops")); err == nil {
		t.Fatal("expected error for nil event")
	}
}

func TestHandle_BelowMaxAttempts_ReturnsSendErr(t *testing.T) {
	fb := &stubSink{}
	h, _ := New(Config{MaxAttempts: 3, Fallback: fb})
	event := sampleEvent()
	origErr := errors.New("transient")

	err := h.Handle(context.Background(), event, origErr)
	if !errors.Is(err, origErr) {
		t.Fatalf("expected original error, got %v", err)
	}
	if len(fb.received) != 0 {
		t.Fatal("fallback should not have been called yet")
	}
}

func TestHandle_ReachesMaxAttempts_ForwardsToFallback(t *testing.T) {
	fb := &stubSink{}
	h, _ := New(Config{MaxAttempts: 2, Fallback: fb})
	event := sampleEvent()
	origErr := errors.New("persistent")

	// first attempt – below threshold
	_ = h.Handle(context.Background(), event, origErr)
	// second attempt – should forward
	err := h.Handle(context.Background(), event, origErr)
	if err != nil {
		t.Fatalf("expected nil after fallback, got %v", err)
	}
	if len(fb.received) != 1 || fb.received[0] != event {
		t.Fatal("expected event forwarded to fallback")
	}
}

func TestHandle_AfterFallback_CounterReset(t *testing.T) {
	fb := &stubSink{}
	h, _ := New(Config{MaxAttempts: 2, Fallback: fb})
	event := sampleEvent()
	origErr := errors.New("err")

	_ = h.Handle(context.Background(), event, origErr)
	_ = h.Handle(context.Background(), event, origErr) // triggers fallback

	key := eventKey(event)
	if h.Attempts(key) != 0 {
		t.Fatalf("expected counter reset to 0, got %d", h.Attempts(key))
	}
}

func TestHandle_FallbackError_Propagated(t *testing.T) {
	fbErr := errors.New("dlq unavailable")
	fb := &stubSink{err: fbErr}
	h, _ := New(Config{MaxAttempts: 1, Fallback: fb})

	err := h.Handle(context.Background(), sampleEvent(), errors.New("send err"))
	if err == nil {
		t.Fatal("expected fallback error to be propagated")
	}
	if !errors.Is(err, fbErr) {
		t.Fatalf("expected wrapped dlq error, got %v", err)
	}
}
