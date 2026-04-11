package eventlog

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/pgstream/pgstream/internal/wal"
)

func sampleEvent() *wal.Event {
	return &wal.Event{
		LSN:       "0/1A2B3C",
		Table:     "public.orders",
		Operation: "INSERT",
	}
}

func TestNew_Stdout(t *testing.T) {
	l, err := New(Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
	_ = l.Close()
}

func TestNew_FileCreated(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	l, err := New(Config{Path: path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer l.Close()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist: %v", err)
	}
}

func TestRecord_WritesEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	l, err := New(Config{Path: path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer l.Close()

	if err := l.Record(sampleEvent(), "sent", ""); err != nil {
		t.Fatalf("Record error: %v", err)
	}

	f, _ := os.Open(path)
	defer f.Close()
	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		t.Fatal("expected at least one line")
	}
	var entry Entry
	if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if entry.Table != "public.orders" {
		t.Errorf("expected table public.orders, got %s", entry.Table)
	}
	if entry.Status != "sent" {
		t.Errorf("expected status sent, got %s", entry.Status)
	}
	if entry.Error != "" {
		t.Errorf("expected empty error, got %s", entry.Error)
	}
}

func TestRecord_WithError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	l, _ := New(Config{Path: path})
	defer l.Close()

	_ = l.Record(sampleEvent(), "failed", "sink unavailable")

	f, _ := os.Open(path)
	defer f.Close()
	var entry Entry
	scanner := bufio.NewScanner(f)
	scanner.Scan()
	_ = json.Unmarshal(scanner.Bytes(), &entry)
	if entry.Error != "sink unavailable" {
		t.Errorf("expected error message, got %q", entry.Error)
	}
}

func TestRecord_NilEvent_NoError(t *testing.T) {
	l, _ := New(Config{})
	defer l.Close()
	if err := l.Record(nil, "sent", ""); err != nil {
		t.Fatalf("expected no error for nil event, got %v", err)
	}
}
