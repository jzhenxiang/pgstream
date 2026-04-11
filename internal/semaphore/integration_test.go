package semaphore

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
)

func TestConcurrentAcquireRelease_NoDataRace(t *testing.T) {
	const (
		workers  = 50
		max      = 10
	)

	sem, err := New(max)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var (
		wg      sync.WaitGroup
		peak    int64
		current int64
	)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := sem.Acquire(context.Background()); err != nil {
				return
			}
			now := atomic.AddInt64(&current, 1)
			for {
				old := atomic.LoadInt64(&peak)
				if now <= old || atomic.CompareAndSwapInt64(&peak, old, now) {
					break
				}
			}
			atomic.AddInt64(&current, -1)
			sem.Release()
		}()
	}

	wg.Wait()

	if p := atomic.LoadInt64(&peak); p > int64(max) {
		t.Errorf("peak concurrency %d exceeded max %d", p, max)
	}
	if sem.Current() != 0 {
		t.Errorf("expected current=0 after all goroutines done, got %d", sem.Current())
	}
}
