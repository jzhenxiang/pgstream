package heartbeat

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// mockSender counts how many times SendStandbyStatus is called.
type mockSender struct {
	calls atomic.Int64
	err   error
}

func (m *mockSender) SendStandbyStatus(_ context.Context) error {
	m.calls.Add(1)
	return m.err
}

func TestNew_NilSender_ReturnsError(t *testing.T) {
	_, err := New(nil, Config{})
	if err == nil {
		t.Fatal("expected error for nil sender")
	}
}

func TestNew_DefaultInterval(t *testing.T) {
	hb, err := New(&mockSender{}, Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hb.cfg.Interval != DefaultInterval {
		t.Errorf("want %v, got %v", DefaultInterval, hb.cfg.Interval)
	}
}

func TestNew_CustomInterval(t *testing.T) {
	hb, err := New(&mockSender{}, Config{Interval: 3 * time.Second})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hb.cfg.Interval != 3*time.Second {
		t.Errorf("want 3s, got %v", hb.cfg.Interval)
	}
}

func TestRun_StopsOnContextCancel(t *testing.T) {
	sender := &mockSender{}
	hb, _ := New(sender, Config{Interval: 50 * time.Millisecond})

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- hb.Run(ctx) }()

	time.Sleep(160 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not stop after context cancellation")
	}

	if sender.calls.Load() < 2 {
		t.Errorf("expected at least 2 heartbeats, got %d", sender.calls.Load())
	}
}

func TestRun_SenderError_ContinuesRunning(t *testing.T) {
	sender := &mockSender{err: errors.New("connection reset")}
	hb, _ := New(sender, Config{Interval: 30 * time.Millisecond})

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	// Should not return early despite sender errors.
	err := hb.Run(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
	if sender.calls.Load() < 2 {
		t.Errorf("expected at least 2 calls, got %d", sender.calls.Load())
	}
}
