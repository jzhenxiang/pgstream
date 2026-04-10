package backpressure_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/your-org/pgstream/internal/backpressure"
)

// TestConcurrentAcquireRelease verifies the controller is safe under concurrent
// access from multiple goroutines.
func TestConcurrentAcquireRelease(t *testing.T) {
	const workers = 8
	const slots = 4

	ctrl := backpressure.New(backpressure.Config{
		MaxPending:     slots,
		AcquireTimeout: 2 * time.Second,
	})

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := context.Background()
			if err := ctrl.Acquire(ctx); err != nil {
				// Some goroutines may time out when slots are full; that is
				// expected behaviour — just return.
				return
			}
			time.Sleep(10 * time.Millisecond)
			ctrl.Release()
		}()
	}
	wg.Wait()

	// After all goroutines finish pending should be 0.
	if p := ctrl.Pending(); p != 0 {
		t.Fatalf("expected 0 pending after concurrent run, got %d", p)
	}
}
