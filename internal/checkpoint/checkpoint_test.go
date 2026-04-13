package checkpoint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew_CreatesEmptyCheckpoint(t *testing.T) {
	dir := t.TempDir()
	cp, err := New(filepath.Join(dir, "wal.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cp.Get() != "" {
		t.Errorf("expected empty LSN, got %q", cp.Get())
	}
}

func TestNew_LoadsExistingCheckpoint(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "wal.json")

	cp1, _ := New(path)
	if err := cp1.Save("0/1A2B3C"); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	cp2, err := New(path)
	if err != nil {
		t.Fatalf("unexpected error on reload: %v", err)
	}
	if cp2.Get() != "0/1A2B3C" {
		t.Errorf("expected LSN %q, got %q", "0/1A2B3C", cp2.Get())
	}
}

func TestSave_PersistsToDisk(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "wal.json")
	cp, _ := New(path)

	if err := cp.Save("0/DEADBEEF"); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected checkpoint file to exist after Save")
	}
	if cp.Get() != "0/DEADBEEF" {
		t.Errorf("expected %q, got %q", "0/DEADBEEF", cp.Get())
	}
}

func TestNew_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "wal.json")
	if err := os.WriteFile(path, []byte("not-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := New(path)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestSave_OverwritesPreviousValue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "wal.json")
	cp, _ := New(path)

	if err := cp.Save("0/AABBCC"); err != nil {
		t.Fatalf("first save failed: %v", err)
	}
	if err := cp.Save("0/112233"); err != nil {
		t.Fatalf("second save failed: %v", err)
	}

	// Reload from disk to confirm the latest value was persisted.
	cp2, err := New(path)
	if err != nil {
		t.Fatalf("unexpected error on reload: %v", err)
	}
	if cp2.Get() != "0/112233" {
		t.Errorf("expected %q after overwrite, got %q", "0/112233", cp2.Get())
	}
}
