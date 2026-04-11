// Package eventlog provides a structured audit log for WAL events
// processed by pgstream, recording key lifecycle transitions.
package eventlog

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/pgstream/pgstream/internal/wal"
)

// Entry represents a single audit log record.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	LSN       string    `json:"lsn"`
	Table     string    `json:"table"`
	Operation string    `json:"operation"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
}

// Logger writes event entries to a file or stdout.
type Logger struct {
	mu   sync.Mutex
	enc  *json.Encoder
	f    *os.File
}

// Config holds configuration for the event logger.
type Config struct {
	// Path is the file path to write logs. Empty means stdout.
	Path string
}

// New creates a new Logger. If cfg.Path is empty, logs are written to stdout.
func New(cfg Config) (*Logger, error) {
	var (
		f   *os.File
		err error
	)
	if cfg.Path != "" {
		f, err = os.OpenFile(cfg.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, fmt.Errorf("eventlog: open file: %w", err)
		}
	} else {
		f = os.Stdout
	}
	return &Logger{
		enc: json.NewEncoder(f),
		f:   f,
	}, nil
}

// Record writes an entry derived from a WAL event and a status string.
func (l *Logger) Record(ev *wal.Event, status, errMsg string) error {
	if ev == nil {
		return nil
	}
	e := Entry{
		Timestamp: time.Now().UTC(),
		LSN:       ev.LSN,
		Table:     ev.Table,
		Operation: ev.Operation,
		Status:    status,
		Error:     errMsg,
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.enc.Encode(e)
}

// Close releases resources held by the logger.
func (l *Logger) Close() error {
	if l.f != nil && l.f != os.Stdout {
		return l.f.Close()
	}
	return nil
}
