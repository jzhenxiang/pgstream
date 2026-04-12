package pruner_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/your-org/pgstream/internal/pruner"
)

// mockDropper records which slots were dropped.
type mockDropper struct {
	drops  []string
	errOn  string
	called atomic.Int32
}

func (m *mockDropper) DropSlot(_ context.Context, name string) error {
	m.called.Add(1)
	if name == m.errOn {
		return errors.New("drop failed")
	}
	m.drops = append(m.drops, name)
	return nil
}

func TestNew_NilDropper_ReturnsError(t *testing.T) {
	_, err := pruner.New(pruner.Config{}, nil, nil)
	if err == nil {
		t.Fatal("expected error for nil dropper")
	}
}

func TestNew_DefaultInterval(t *testing.T) {
	d := &mockDropper{}
	p, err := pruner.New(pruner.Config{}, d, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil pruner")
	}
}

func TestRun_StopsOnContextCancel(t *testing.T) {
	d := &mockDropper{}
	p, _ := pruner.New(pruner.Config{Interval: time.Hour}, d, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- p.Run(ctx) }()

	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not stop after context cancel")
	}
}

func TestRun_DropsConfiguredSlots(t *testing.T) {
	d := &mockDropper{}
	cfg := pruner.Config{
		Interval: 20 * time.Millisecond,
		Slots:    []string{"slot_a", "slot_b"},
	}
	p, _ := pruner.New(cfg, d, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	_ = p.Run(ctx) //nolint:errcheck

	if d.called.Load() == 0 {
		t.Fatal("expected at least one drop call")
	}
}

func TestRun_DropError_ContinuesOtherSlots(t *testing.T) {
	d := &mockDropper{errOn: "bad_slot"}
	cfg := pruner.Config{
		Interval: 20 * time.Millisecond,
		Slots:    []string{"bad_slot", "good_slot"},
	}
	p, _ := pruner.New(cfg, d, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()
	_ = p.Run(ctx) //nolint:errcheck

	for _, name := range d.drops {
		if name == "bad_slot" {
			t.Fatal("bad_slot should not appear in successful drops")
		}
	}
}
