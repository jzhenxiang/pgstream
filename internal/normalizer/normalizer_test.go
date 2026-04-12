package normalizer

import (
	"testing"

	"github.com/pgstream/pgstream/internal/wal"
)

func baseEvent() *wal.Event {
	return &wal.Event{
		Table: "users",
		Data: map[string]any{
			"email": "  User@Example.COM  ",
			"name":  "  Alice  ",
			"score": 42,
		},
	}
}

func TestNew_EmptyConfig(t *testing.T) {
	n := New(nil)
	if n == nil {
		t.Fatal("expected non-nil normalizer")
	}
	if len(n.rules) != 0 {
		t.Errorf("expected 0 rules, got %d", len(n.rules))
	}
}

func TestApply_NilEvent_ReturnsNil(t *testing.T) {
	n := New(&Config{Rules: []Rule{{Table: "users", Column: "email", Mode: "lowercase"}}})
	if got := n.Apply(nil); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestApply_NoRules_ReturnsSamePointer(t *testing.T) {
	n := New(nil)
	ev := baseEvent()
	got := n.Apply(ev)
	if got != ev {
		t.Error("expected same pointer when no rules configured")
	}
}

func TestApply_Lowercase(t *testing.T) {
	n := New(&Config{Rules: []Rule{{Table: "users", Column: "email", Mode: "lowercase"}}})
	got := n.Apply(baseEvent())
	if got.Data["email"] != "  user@example.com  " {
		t.Errorf("unexpected email: %v", got.Data["email"])
	}
}

func TestApply_Trim(t *testing.T) {
	n := New(&Config{Rules: []Rule{{Table: "users", Column: "name", Mode: "trim"}}})
	got := n.Apply(baseEvent())
	if got.Data["name"] != "Alice" {
		t.Errorf("unexpected name: %v", got.Data["name"])
	}
}

func TestApply_Uppercase(t *testing.T) {
	n := New(&Config{Rules: []Rule{{Table: "users", Column: "name", Mode: "uppercase"}}})
	got := n.Apply(baseEvent())
	if got.Data["name"] != "  ALICE  " {
		t.Errorf("unexpected name: %v", got.Data["name"])
	}
}

func TestApply_NonStringField_IsUntouched(t *testing.T) {
	n := New(&Config{Rules: []Rule{{Table: "users", Column: "score", Mode: "lowercase"}}})
	got := n.Apply(baseEvent())
	if got.Data["score"] != 42 {
		t.Errorf("expected score 42, got %v", got.Data["score"])
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	n := New(&Config{Rules: []Rule{{Table: "users", Column: "email", Mode: "lowercase"}}})
	original := baseEvent()
	orig := original.Data["email"]
	n.Apply(original)
	if original.Data["email"] != orig {
		t.Error("original event was mutated")
	}
}

func TestApply_WildcardTable(t *testing.T) {
	n := New(&Config{Rules: []Rule{{Table: "*", Column: "email", Mode: "lowercase"}}})
	ev := &wal.Event{Table: "orders", Data: map[string]any{"email": "ADMIN@EXAMPLE.COM"}}
	got := n.Apply(ev)
	if got.Data["email"] != "admin@example.com" {
		t.Errorf("unexpected email: %v", got.Data["email"])
	}
}

func TestApply_TableMismatch_SkipsRule(t *testing.T) {
	n := New(&Config{Rules: []Rule{{Table: "orders", Column: "email", Mode: "lowercase"}}})
	got := n.Apply(baseEvent()) // table is "users"
	if got.Data["email"] != "  User@Example.COM  " {
		t.Errorf("expected unchanged email, got %v", got.Data["email"])
	}
}
