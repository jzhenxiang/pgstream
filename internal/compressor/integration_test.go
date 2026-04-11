package compressor_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"sync"
	"testing"

	"github.com/your-org/pgstream/internal/compressor"
)

// TestConcurrentCompress_NoDataRace verifies that a single Compressor instance
// is safe for concurrent use across multiple goroutines.
func TestConcurrentCompress_NoDataRace(t *testing.T) {
	c, err := compressor.New(compressor.Config{Algorithm: compressor.Gzip})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	const workers = 20
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			src := []byte("concurrent pgstream payload for compression test")
			out, err := c.Compress(src)
			if err != nil {
				t.Errorf("Compress: %v", err)
				return
			}
			r, err := gzip.NewReader(bytes.NewReader(out))
			if err != nil {
				t.Errorf("gzip.NewReader: %v", err)
				return
			}
			decompressed, err := io.ReadAll(r)
			if err != nil {
				t.Errorf("ReadAll: %v", err)
				return
			}
			if !bytes.Equal(src, decompressed) {
				t.Errorf("round-trip mismatch")
			}
		}()
	}
	wg.Wait()
}
