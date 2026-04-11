package circuitbreaker

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.MaxFailures != 5 {
		t.Fatalf("expected MaxFailures 5, got %d", cfg.MaxFailures)
	}
	if cfg.ResetTimeout != 30*time.Second {
		t.Fatalf("expected ResetTimeout 30s, got %v", cfg.ResetTimeout)
	}
}

func TestNew_ZeroValues_UsesDefaults(t *testing.T) {
	cb := New(Config{})
	if cb.cfg.MaxFailures != 5 {
		t.Fatalf("expected default MaxFailures 5, got %d", cb.cfg.MaxFailures)
	}
	if cb.cfg.ResetTimeout != 30*time.Second {
		t.Fatalf("expected default ResetTimeout 30s, got %v", cb.cfg.ResetTimeout)
	}
}

func TestAllow_InitiallyClosed(t *testing.T) {
	cb := New(DefaultConfig())
	if err := cb.Allow(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRecordFailure_OpensCircuit(t *testing.T) {
	cb := New(Config{MaxFailures: 3, ResetTimeout: time.Minute})
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	if cb.State() != StateOpen {
		t.Fatalf("expected StateOpen, got %v", cb.State())
	}
	if err := cb.Allow(); err != ErrOpen {
		t.Fatalf("expected ErrOpen, got %v", err)
	}
}

func TestRecordSuccess_ClosesCircuit(t *testing.T) {
	cb := New(Config{MaxFailures: 2, ResetTimeout: time.Minute})
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Fatal("expected circuit to be open")
	}
	// Simulate reset timeout elapsed.
	cb.mu.Lock()
	cb.lastFailure = time.Now().Add(-2 * time.Minute)
	cb.mu.Unlock()

	if err := cb.Allow(); err != nil {
		t.Fatalf("expected nil after timeout, got %v", err)
	}
	if cb.State() != StateHalfOpen {
		t.Fatalf("expected StateHalfOpen, got %v", cb.State())
	}
	cb.RecordSuccess()
	if cb.State() != StateClosed {
		t.Fatalf("expected StateClosed after success, got %v", cb.State())
	}
}

func TestAllow_OpenBeforeTimeout_ReturnsError(t *testing.T) {
	cb := New(Config{MaxFailures: 1, ResetTimeout: time.Hour})
	cb.RecordFailure()
	if err := cb.Allow(); err != ErrOpen {
		t.Fatalf("expected ErrOpen, got %v", err)
	}
}

func TestHalfOpen_FailureReopens(t *testing.T) {
	cb := New(Config{MaxFailures: 1, ResetTimeout: time.Millisecond})
	cb.RecordFailure()
	time.Sleep(5 * time.Millisecond)

	// Transitions to HalfOpen.
	if err := cb.Allow(); err != nil {
		t.Fatalf("expected nil in HalfOpen, got %v", err)
	}
	// Probe fails – should reopen.
	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Fatalf("expected StateOpen after probe failure, got %v", cb.State())
	}
}
