package limiter_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/your-org/pgstream/internal/limiter"
)

// TestConcurrentAllow_NoDataRace verifies that concurrent calls to Allow do not
// trigger the race detector.
func TestConcurrentAllow_NoDataRace(t *testing.T) {
	lim, err := limiter.New(limiter.Config{MaxEventsPerSecond: 500, BurstSize: 500})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	const goroutines = 20
	const callsEach = 50

	var (
		wg      sync.WaitGroup
		allowed atomic.Int64
	)

	ctx := context.Background()
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < callsEach; i++ {
				if lim.Allow(ctx) == nil {
					allowed.Add(1)
				}
			}
		}()
	}
	wg.Wait()

	if allowed.Load() == 0 {
		t.Fatal("expected at least some events to be allowed")
	}
}
