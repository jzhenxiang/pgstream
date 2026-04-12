package window

import (
	"sync"
	"testing"
	"time"
)

func TestConcurrentAdd_NoDataRace(t *testing.T) {
	w, err := New(time.Minute, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	const goroutines = 20
	const addsPerGoroutine = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < addsPerGoroutine; j++ {
				w.Add(1)
				_ = w.Count()
			}
		}()
	}
	wg.Wait()

	expected := int64(goroutines * addsPerGoroutine)
	if c := w.Count(); c != expected {
		t.Fatalf("expected %d, got %d", expected, c)
	}
}
