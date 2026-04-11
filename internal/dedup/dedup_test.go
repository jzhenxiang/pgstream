package dedup_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/your-org/pgstream/internal/dedup"
)

func TestNew_DefaultWindowSize(t *testing.T) {
	d := dedup.New(nil)
	if d.Len() != 0 {
		t.Fatalf("expected empty dedup, got %d", d.Len())
	}
}

func TestIsDuplicate_NewKey_ReturnsFalse(t *testing.T) {
	d := dedup.New(nil)
	if d.IsDuplicate("key1") {
		t.Fatal("expected false for unseen key")
	}
}

func TestIsDuplicate_SeenKey_ReturnsTrue(t *testing.T) {
	d := dedup.New(nil)
	d.IsDuplicate("key1")
	if !d.IsDuplicate("key1") {
		t.Fatal("expected true for already-seen key")
	}
}

func TestIsDuplicate_DifferentKeys(t *testing.T) {
	d := dedup.New(nil)
	d.IsDuplicate("a")
	if d.IsDuplicate("b") {
		t.Fatal("different key should not be a duplicate")
	}
}

func TestWindowEviction(t *testing.T) {
	d := dedup.New(&dedup.Config{WindowSize: 3})
	for i := 0; i < 3; i++ {
		d.IsDuplicate(fmt.Sprintf("k%d", i))
	}
	// adding a 4th key should evict "k0"
	d.IsDuplicate("k3")
	if d.Len() != 3 {
		t.Fatalf("expected window size 3, got %d", d.Len())
	}
	// k0 was evicted, so it should be treated as new
	if d.IsDuplicate("k0") {
		t.Fatal("evicted key should not be a duplicate")
	}
}

func TestTTL_ExpiredKey_TreatedAsNew(t *testing.T) {
	d := dedup.New(&dedup.Config{TTL: 10 * time.Millisecond})
	d.IsDuplicate("key")
	time.Sleep(20 * time.Millisecond)
	if d.IsDuplicate("key") {
		t.Fatal("expired key should not be a duplicate")
	}
}

func TestReset_ClearsAllKeys(t *testing.T) {
	d := dedup.New(nil)
	d.IsDuplicate("a")
	d.IsDuplicate("b")
	d.Reset()
	if d.Len() != 0 {
		t.Fatalf("expected 0 after reset, got %d", d.Len())
	}
	if d.IsDuplicate("a") {
		t.Fatal("key should not be duplicate after reset")
	}
}

func TestConcurrentAccess_NoDataRace(t *testing.T) {
	d := dedup.New(&dedup.Config{WindowSize: 64})
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", n%20)
			d.IsDuplicate(key)
		}(i)
	}
	wg.Wait()
}
