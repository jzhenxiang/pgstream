package buffer_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pgstream/pgstream/internal/buffer"
	"github.com/pgstream/pgstream/internal/wal"
)

func TestConcurrentAdd_NoDataRace(t *testing.T) {
	var total int64
	buf := buffer.New(buffer.Config{
		MaxSize:       10,
		FlushInterval: 50 * time.Millisecond,
	}, func(events []*wal.Event) error {
		atomic.AddInt64(&total, int64(len(events)))
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		defer wg.Done()
		_ = buf.Run(ctx)
	}()

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_ = buf.Add(&wal.Event{Table: "orders", Action: "UPDATE"})
				time.Sleep(time.Millisecond)
			}
		}()
	}

	wg.Wait()
	if atomic.LoadInt64(&total) == 0 {
		t.Error("expected some events to be flushed")
	}
}
