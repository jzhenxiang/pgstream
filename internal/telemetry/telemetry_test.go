package telemetry

import (
	"testing"
	"time"
)

func TestNew_InitialisesZeroCounters(t *testing.T) {
	tel := New()
	snap := tel.Snapshot()

	if snap.EventsReceived != 0 {
		t.Errorf("expected 0 events received, got %d", snap.EventsReceived)
	}
	if snap.EventsProcessed != 0 {
		t.Errorf("expected 0 events processed, got %d", snap.EventsProcessed)
	}
	if snap.EventsFailed != 0 {
		t.Errorf("expected 0 events failed, got %d", snap.EventsFailed)
	}
	if snap.BytesProcessed != 0 {
		t.Errorf("expected 0 bytes processed, got %d", snap.BytesProcessed)
	}
}

func TestIncReceived(t *testing.T) {
	tel := New()
	tel.IncReceived()
	tel.IncReceived()
	if got := tel.Snapshot().EventsReceived; got != 2 {
		t.Errorf("expected 2, got %d", got)
	}
}

func TestIncProcessed(t *testing.T) {
	tel := New()
	tel.IncProcessed()
	if got := tel.Snapshot().EventsProcessed; got != 1 {
		t.Errorf("expected 1, got %d", got)
	}
}

func TestIncFailed(t *testing.T) {
	tel := New()
	tel.IncFailed()
	tel.IncFailed()
	tel.IncFailed()
	if got := tel.Snapshot().EventsFailed; got != 3 {
		t.Errorf("expected 3, got %d", got)
	}
}

func TestIncFiltered(t *testing.T) {
	tel := New()
	tel.IncFiltered()
	if got := tel.Snapshot().EventsFiltered; got != 1 {
		t.Errorf("expected 1, got %d", got)
	}
}

func TestAddBytes(t *testing.T) {
	tel := New()
	tel.AddBytes(512)
	tel.AddBytes(512)
	if got := tel.Snapshot().BytesProcessed; got != 1024 {
		t.Errorf("expected 1024, got %d", got)
	}
}

func TestSnapshot_CapturedAt_IsRecent(t *testing.T) {
	before := time.Now()
	tel := New()
	snap := tel.Snapshot()
	after := time.Now()

	if snap.CapturedAt.Before(before) || snap.CapturedAt.After(after) {
		t.Errorf("CapturedAt %v not in expected range [%v, %v]", snap.CapturedAt, before, after)
	}
}

func TestSnapshot_Uptime_NonEmpty(t *testing.T) {
	tel := New()
	snap := tel.Snapshot()
	if snap.Uptime == "" {
		t.Error("expected non-empty uptime string")
	}
}
