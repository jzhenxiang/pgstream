package acknowledge_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/your-org/pgstream/internal/acknowledge"
)

// mockSender records the last LSN sent and can optionally return an error.
type mockSender struct {
	calls atomic.Int32
	last  atomic.Uint64
	err   error
}

func (m *mockSender) SendStandbyStatusUpdate(_ context.Context, lsn uint64) error {
	if m.err != nil {
		return m.err
	}
	m.calls.Add(1)
	m.last.Store(lsn)
	return nil
}

func TestNew_NilSender_ReturnsError(t *testing.T) {
	_, err := acknowledge.New(nil, acknowledge.Config{})
	if !errors.Is(err, acknowledge.ErrNilSender) {
		t.Fatalf("expected ErrNilSender, got %v", err)
	}
}

func TestNew_DefaultFlushInterval(t *testing.T) {
	ack, err := acknowledge.New(&mockSender{}, acknowledge.Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ack == nil {
		t.Fatal("expected non-nil Acknowledger")
	}
}

func TestTrack_And_Flush_SendsLSN(t *testing.T) {
	sender := &mockSender{}
	ack, _ := acknowledge.New(sender, acknowledge.Config{})

	ack.Track(100)
	ack.Track(200)
	ack.Track(150) // lower than 200, should be ignored

	if err := ack.Flush(context.Background()); err != nil {
		t.Fatalf("unexpected flush error: %v", err)
	}
	if got := sender.last.Load(); got != 200 {
		t.Fatalf("expected LSN 200, got %d", got)
	}
}

func TestFlush_NoAdvance_SkipsSend(t *testing.T) {
	sender := &mockSender{}
	ack, _ := acknowledge.New(sender, acknowledge.Config{})

	// No Track calls — flush should be a no-op.
	if err := ack.Flush(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sender.calls.Load() != 0 {
		t.Fatal("expected zero sender calls")
	}
}

func TestFlush_SenderError_Propagates(t *testing.T) {
	sentinelErr := errors.New("send failed")
	sender := &mockSender{err: sentinelErr}
	ack, _ := acknowledge.New(sender, acknowledge.Config{})

	ack.Track(42)
	err := ack.Flush(context.Background())
	if !errors.Is(err, sentinelErr) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestRun_StopsOnContextCancel(t *testing.T) {
	sender := &mockSender{}
	ack, _ := acknowledge.New(sender, acknowledge.Config{FlushInterval: 100 * time.Millisecond})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- ack.Run(ctx) }()

	ack.Track(99)
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not stop after context cancellation")
	}
}

func TestRun_TicksAndFlushes(t *testing.T) {
	sender := &mockSender{}
	ack, _ := acknowledge.New(sender, acknowledge.Config{FlushInterval: 50 * time.Millisecond})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go ack.Run(ctx) //nolint:errcheck

	ack.Track(77)
	time.Sleep(200 * time.Millisecond)

	if sender.last.Load() != 77 {
		t.Fatalf("expected LSN 77 to be flushed, got %d", sender.last.Load())
	}
}
