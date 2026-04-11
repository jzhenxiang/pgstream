package eventlog_test

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/pgstream/pgstream/internal/eventlog"
	"github.com/pgstream/pgstream/internal/wal"
)

func TestConcurrentRecord_NoDataRace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "concurrent.jsonl")
	l, err := eventlog.New(eventlog.Config{Path: path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
 l.Close()

	const workers = 20
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			ev := &wal.Event{LSN: "0/AABBCC", Table: "public.items", Operation: "UPDATE"}
			_ = l.Record(ev, "sent", "")
		}()
	}
	wg.Wait()

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open file: %v", err)
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var e eventlog.Entry
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			t.Errorf("invalid JSON line: %v", err)
		}
		count++
	}
	if count != workers {
		t.Errorf("expected %d entries, got %d", workers, count)
	}
}
