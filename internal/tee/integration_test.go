package tee_test

import (
	"context"
	"sync"
	"testing"

	"pgstream/internal/tee"
	"pgstream/internal/wal"
)

type safeSink struct {
	mu    sync.Mutex
	count int
}

func (s *safeSink) Send(_ context.Context, _ *wal.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.count++
	return nil
}

func TestConcurrentSend_NoDataRace(t *testing.T) {
	a, b := &safeSink{}, &safeSink{}
	tee, err := tee.New(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_ = tee.Send(context.Background(), &wal.Event{})
		}()
	}
	wg.Wait()

	if a.count != goroutines || b.count != goroutines {
		t.Fatalf("expected %d calls each, got a=%d b=%d", goroutines, a.count, b.count)
	}
}
