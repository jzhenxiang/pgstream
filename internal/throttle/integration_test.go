package throttle_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/pgstream/pgstream/internal/throttle"
)

// TestConcurrentIncreaseDecrease_NoDataRace ensures Increase/Decrease/Current
// are safe to call from multiple goroutines simultaneously.
func TestConcurrentIncreaseDecrease_NoDataRace(t *testing.T) {
	th, err := throttle.New(throttle.DefaultConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	const workers = 20
	var wg sync.WaitGroup
	wg.Add(workers)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				th.Increase()
			} else {
				th.Decrease()
			}
			_ = th.Current()
			// Wait with a very short context so the test stays fast.
			shortCtx, shortCancel := context.WithTimeout(ctx, 5*time.Millisecond)
			defer shortCancel()
			_ = th.Wait(shortCtx)
		}(i)
	}

	wg.Wait()
}
