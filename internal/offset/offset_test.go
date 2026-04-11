package offset

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func tmpFile(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "offset.json")
}

func TestNew_StartsAtZero(t *testing.T) {
	tr, err := New(tmpFile(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.Current() != 0 {
		t.Fatalf("expected 0, got %d", tr.Current())
	}
}

func TestCommit_UpdatesCurrent(t *testing.T) {
	tr, _ := New(tmpFile(t))
	if err := tr.Commit(42); err != nil {
		t.Fatalf("commit error: %v", err)
	}
	if tr.Current() != 42 {
		t.Fatalf("expected 42, got %d", tr.Current())
	}
}

func TestCommit_PersistsToDisk(t *testing.T) {
	path := tmpFile(t)
	tr, _ := New(path)
	_ = tr.Commit(99)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	var s persistedState
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if s.Position != 99 {
		t.Fatalf("expected 99, got %d", s.Position)
	}
}

func TestNew_LoadsExistingPosition(t *testing.T) {
	path := tmpFile(t)
	tr, _ := New(path)
	_ = tr.Commit(1234)

	tr2, err := New(path)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}
	if tr2.Current() != 1234 {
		t.Fatalf("expected 1234, got %d", tr2.Current())
	}
}

func TestNew_InvalidJSON_ReturnsError(t *testing.T) {
	path := tmpFile(t)
	_ = os.WriteFile(path, []byte("not-json"), 0o644)
	_, err := New(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestCommit_MultipleCommits_LastWins(t *testing.T) {
	path := tmpFile(t)
	tr, _ := New(path)
	for _, pos := range []Position{10, 20, 30} {
		if err := tr.Commit(pos); err != nil {
			t.Fatalf("commit %d: %v", pos, err)
		}
	}
	if tr.Current() != 30 {
		t.Fatalf("expected 30, got %d", tr.Current())
	}
}
