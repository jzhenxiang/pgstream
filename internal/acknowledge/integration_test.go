package acknowledge_test

import (
	"context"
	"sync"
	"testing"

	"github.com/your-org/pgstream/internal/acknowledge"
)

// TestConcurrentTrack_NoDataRace verifies that concurrent calls to Track and
// Flush do not produce data races (run with -race).
func TestConcurrentTrack_NoDataRace(t *testing.T) {
	sender := &mockSender{}
	ack, err := acknowledge.New(sender, acknowledge.Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(lsn uint64) {
			defer wg.Done()
			ack.Track(lsn)
			_ = ack.Flush(context.Background())
		}(uint64(i * 10))
	}

	wg.Wait()
}
