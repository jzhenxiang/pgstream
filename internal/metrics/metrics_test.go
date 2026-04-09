package metrics

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	before := time.Now()
	m := New()
	after := time.Now()

	if m.StartTime.Before(before) || m.StartTime.After(after) {
		t.Errorf("expected StartTime between %v and %v, got %v", before, after, m.StartTime)
	}
}

func TestRecordReceived(t *testing.T) {
	m := New()
	m.RecordReceived()
	m.RecordReceived()

	if got := m.MessagesReceived.Load(); got != 2 {
		t.Errorf("expected 2 received, got %d", got)
	}
}

func TestRecordProcessed(t *testing.T) {
	m := New()
	m.RecordProcessed(128)
	m.RecordProcessed(256)

	if got := m.MessagesProcessed.Load(); got != 2 {
		t.Errorf("expected 2 processed, got %d", got)
	}
	if got := m.BytesProcessed.Load(); got != 384 {
		t.Errorf("expected 384 bytes, got %d", got)
	}
}

func TestRecordFailed(t *testing.T) {
	m := New()
	m.RecordFailed()

	if got := m.MessagesFailed.Load(); got != 1 {
		t.Errorf("expected 1 failed, got %d", got)
	}
}

func TestSnapshot(t *testing.T) {
	m := New()
	m.RecordReceived()
	m.RecordReceived()
	m.RecordProcessed(100)
	m.RecordFailed()

	snap := m.Snapshot()

	if snap.MessagesReceived != 2 {
		t.Errorf("expected 2 received, got %d", snap.MessagesReceived)
	}
	if snap.MessagesProcessed != 1 {
		t.Errorf("expected 1 processed, got %d", snap.MessagesProcessed)
	}
	if snap.MessagesFailed != 1 {
		t.Errorf("expected 1 failed, got %d", snap.MessagesFailed)
	}
	if snap.BytesProcessed != 100 {
		t.Errorf("expected 100 bytes, got %d", snap.BytesProcessed)
	}
	if snap.UptimeSeconds < 0 {
		t.Errorf("expected non-negative uptime, got %d", snap.UptimeSeconds)
	}
}
