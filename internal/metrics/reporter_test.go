package metrics

import (
	"context"
	"testing"
	"time"
)

func TestNewReporter_DefaultInterval(t *testing.T) {
	m := New()
	r := NewReporter(m, 0)

	if r.interval != 30*time.Second {
		t.Errorf("expected default interval 30s, got %v", r.interval)
	}
}

func TestNewReporter_CustomInterval(t *testing.T) {
	m := New()
	r := NewReporter(m, 5*time.Second)

	if r.interval != 5*time.Second {
		t.Errorf("expected interval 5s, got %v", r.interval)
	}
}

func TestReporter_Run_StopsOnContextCancel(t *testing.T) {
	m := New()
	m.RecordReceived()
	m.RecordProcessed(64)

	r := NewReporter(m, 10*time.Second)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		r.Run(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Error("reporter did not stop after context cancellation")
	}
}

func TestReporter_Run_TicksAndReports(t *testing.T) {
	m := New()
	m.RecordReceived()

	r := NewReporter(m, 50*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		r.Run(ctx)
		close(done)
	}()

	select {
	case <-done:
		// reporter exited cleanly after context timeout
	case <-time.After(500 * time.Millisecond):
		t.Error("reporter did not exit in time")
	}
}
