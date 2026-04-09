package checkpoint

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// Checkpoint tracks the last successfully processed WAL LSN position.
type Checkpoint struct {
	mu       sync.RWMutex
	filePath string
	LSN      string `json:"lsn"`
}

// New creates a new Checkpoint backed by the given file path.
// If the file exists, it loads the last saved LSN.
func New(filePath string) (*Checkpoint, error) {
	cp := &Checkpoint{filePath: filePath}
	if err := cp.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("checkpoint: load: %w", err)
	}
	return cp, nil
}

// Save persists the current LSN to disk.
func (c *Checkpoint) Save(lsn string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LSN = lsn
	return c.flush()
}

// Get returns the last saved LSN.
func (c *Checkpoint) Get() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.LSN
}

func (c *Checkpoint) load() error {
	data, err := os.ReadFile(c.filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, c)
}

func (c *Checkpoint) flush() error {
	data, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("checkpoint: marshal: %w", err)
	}
	return os.WriteFile(c.filePath, data, 0o644)
}
