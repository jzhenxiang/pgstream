package mapper

import (
	"testing"

	"github.com/pgstream/pgstream/internal/wal"
)

func baseEvent() *wal.Event {
	return &wal.Event{
		Table: "orders",
		Data:  map[string]any{"order_id": 1, "cust_id": 2, "amount": 99.9},
	}
}

func TestNew_EmptyConfig(t *testing.T) {
	m, err := New(Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m == nil {
		t.Fatal("expected non-nil mapper")
	}
}

func TestNew_BlankFieldMapping_ReturnsError(t *testing.T) {
	cfg := Config{
		Rules: map[string]map[string]string{
			"orders": {"": "new_key"},
		},
	}
	_, err := New(cfg)
	if err == nil {
		t.Fatal("expected error for blank source key")
	}
}

func TestApply_NilEvent_ReturnsNil(t *testing.T) {
	m, _ := New(Config{})
	if got := m.Apply(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestApply_NoMatchingRule_ReturnsSamePointer(t *testing.T) {
	m, _ := New(Config{Rules: map[string]map[string]string{"other": {"a": "b"}}})
	ev := baseEvent()
	got := m.Apply(ev)
	if got != ev {
		t.Fatal("expected same pointer when no rule matches")
	}
}

func TestApply_RenamesFields(t *testing.T) {
	cfg := Config{
		Rules: map[string]map[string]string{
			"orders": {"order_id": "id", "cust_id": "customer_id"},
		},
	}
	m, _ := New(cfg)
	ev := baseEvent()
	got := m.Apply(ev)

	if got == ev {
		t.Fatal("expected a new copy, not the same pointer")
	}
	if _, exists := got.Data["order_id"]; exists {
		t.Error("old key 'order_id' should have been removed")
	}
	if got.Data["id"] != 1 {
		t.Errorf("expected id=1, got %v", got.Data["id"])
	}
	if got.Data["customer_id"] != 2 {
		t.Errorf("expected customer_id=2, got %v", got.Data["customer_id"])
	}
	if got.Data["amount"] != 99.9 {
		t.Errorf("expected amount=99.9, got %v", got.Data["amount"])
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	cfg := Config{
		Rules: map[string]map[string]string{
			"orders": {"order_id": "id"},
		},
	}
	m, _ := New(cfg)
	ev := baseEvent()
	m.Apply(ev)

	if _, exists := ev.Data["order_id"]; !exists {
		t.Error("original event should not have been mutated")
	}
}
