package cursor

import (
	"sync"
	"testing"
)

func TestConcurrentAdvance_NoDataRace(t *testing.T) {
	c := New(0)
	const goroutines = 50
	const advances = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(base int) {
			defer wg.Done()
			for j := 0; j < advances; j++ {
				lsn := LSN(base*advances + j)
				c.Advance(lsn)
				_ = c.Current()
				_ = c.HighWaterMark()
			}
		}(i)
	}

	wg.Wait()

	if c.HighWaterMark() == 0 {
		t.Fatal("expected high-water mark to be non-zero after concurrent advances")
	}
}
