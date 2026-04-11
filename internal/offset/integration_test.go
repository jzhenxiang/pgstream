package offset_test

import (
	"path/filepath"
	"sync"
	"testing"

	"github.com/your-org/pgstream/internal/offset"
)

// TestConcurrentCommit_NoDataRace verifies that concurrent Commit and Current
// calls do not trigger the race detector.
func TestConcurrentCommit_NoDataRace(t *testing.T) {
	path := filepath.Join(t.TempDir(), "offset.json")
	tr, err := offset.New(path)
	if err != nil {
		t.Fatalf("new tracker: %v", err)
	}

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			if n%2 == 0 {
				_ = tr.Commit(offset.Position(n))
			} else {
				_ = tr.Current()
			}
		}(i)
	}

	wg.Wait()
}
