// Package offset tracks the last successfully processed WAL LSN position
// and persists it to disk so that pgstream can resume from the correct
// position after a restart.
package offset

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// Position represents a WAL Log Sequence Number as a uint64.
type Position uint64

// Tracker persists and retrieves the last committed WAL offset.
type Tracker struct {
	mu       sync.RWMutex
	current  Position
	filePath string
}

type persistedState struct {
	Position uint64 `json:"position"`
}

// New creates a Tracker. If filePath points to an existing file the stored
// position is loaded; otherwise the tracker starts at zero.
func New(filePath string) (*Tracker, error) {
	t := &Tracker{filePath: filePath}
	if _, err := os.Stat(filePath); err == nil {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("offset: read file: %w", err)
		}
		var s persistedState
		if err := json.Unmarshal(data, &s); err != nil {
			return nil, fmt.Errorf("offset: unmarshal: %w", err)
		}
		t.current = Position(s.Position)
	}
	return t, nil
}

// Commit updates the in-memory position and flushes it to disk.
func (t *Tracker) Commit(pos Position) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.current = pos
	data, err := json.Marshal(persistedState{Position: uint64(pos)})
	if err != nil {
		return fmt.Errorf("offset: marshal: %w", err)
	}
	if err := os.WriteFile(t.filePath, data, 0o644); err != nil {
		return fmt.Errorf("offset: write file: %w", err)
	}
	return nil
}

// Current returns the last committed position.
func (t *Tracker) Current() Position {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.current
}
