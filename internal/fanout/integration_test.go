package fanout

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/pgstream/pgstream/internal/wal"
)

// TestConcurrentSend_NoDataRace exercises the fanout under concurrent senders
// to verify there are no data races. Run with: go test -race ./internal/fanout/...
func TestConcurrentSend_NoDataRace(t *testing.T) {
	const goroutines = 20

	var total atomic.Int64

	makeSink := func() *mockSink {
		return &mockSink{
			sendFn: func(_ context.Context, _ *wal.Event) error {
				total.Add(1)
				return nil
			},
		}
	}

	f, err := New(makeSink(), makeSink(), makeSink())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	done := make(chan struct{})
	for i := 0; i < goroutines; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			_ = f.Send(context.Background(), &wal.Event{})
		}()
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}

	// 3 sinks × goroutines sends
	expected := int64(3 * goroutines)
	if got := total.Load(); got != expected {
		t.Errorf("expected %d total sink calls, got %d", expected, got)
	}
}
