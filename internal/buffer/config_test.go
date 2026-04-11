package buffer

import (
	"testing"
	"time"

	"github.com/pgstream/pgstream/internal/wal"
)

func TestNew_CustomConfig(t *testing.T) {
	cfg := Config{
		MaxSize:       25,
		FlushInterval: 3 * time.Second,
	}
	buf := New(cfg, func(_ []*wal.Event) error { return nil })
	if buf.cfg.MaxSize != 25 {
		t.Errorf("expected MaxSize 25, got %d", buf.cfg.MaxSize)
	}
	if buf.cfg.FlushInterval != 3*time.Second {
		t.Errorf("expected FlushInterval 3s, got %v", buf.cfg.FlushInterval)
	}
}

func TestNew_ZeroMaxSize_UsesDefault(t *testing.T) {
	buf := New(Config{MaxSize: 0}, func(_ []*wal.Event) error { return nil })
	if buf.cfg.MaxSize != DefaultMaxSize {
		t.Errorf("expected default MaxSize %d, got %d", DefaultMaxSize, buf.cfg.MaxSize)
	}
}

func TestNew_ZeroFlushInterval_UsesDefault(t *testing.T) {
	buf := New(Config{FlushInterval: 0}, func(_ []*wal.Event) error { return nil })
	if buf.cfg.FlushInterval != DefaultFlushInterval {
		t.Errorf("expected default FlushInterval %v, got %v", DefaultFlushInterval, buf.cfg.FlushInterval)
	}
}

func TestLen_InitiallyZero(t *testing.T) {
	buf := New(Config{}, func(_ []*wal.Event) error { return nil })
	if buf.Len() != 0 {
		t.Errorf("expected Len 0, got %d", buf.Len())
	}
}

func TestLen_AfterAdd(t *testing.T) {
	buf := New(Config{MaxSize: 50}, func(_ []*wal.Event) error { return nil })
	_ = buf.Add(&wal.Event{})
	_ = buf.Add(&wal.Event{})
	if buf.Len() != 2 {
		t.Errorf("expected Len 2, got %d", buf.Len())
	}
}
