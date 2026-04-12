package redactor

import (
	"testing"

	"github.com/pgstream/pgstream/internal/wal"
)

func baseEvent() *wal.Event {
	return &wal.Event{
		Table: "users",
		Data: map[string]any{
			"id":    1,
			"email": "alice@example.com",
			"phone": "555-1234",
			"name":  "Alice",
		},
	}
}

func TestNew_EmptyConfig(t *testing.T) {
	r := New(Config{})
	if r == nil {
		t.Fatal("expected non-nil redactor")
	}
}

func TestApply_NilEvent_ReturnsNil(t *testing.T) {
	r := New(Config{})
	if got := r.Apply(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestApply_NoRules_ReturnsSamePointer(t *testing.T) {
	r := New(Config{})
	ev := baseEvent()
	got := r.Apply(ev)
	if got != ev {
		t.Fatal("expected same pointer when no rules configured")
	}
}

func TestApply_BlankStrategy(t *testing.T) {
	r := New(Config{Rules: []Rule{
		{Table: "users", Columns: []string{"email"}, Strategy: StrategyBlank},
	}})
	got := r.Apply(baseEvent())
	if got.Data["email"] != "" {
		t.Fatalf("expected blank email, got %v", got.Data["email"])
	}
	if got.Data["name"] != "Alice" {
		t.Fatal("non-redacted field should be unchanged")
	}
}

func TestApply_HashStrategy(t *testing.T) {
	r := New(Config{Rules: []Rule{
		{Table: "users", Columns: []string{"email"}, Strategy: StrategyHash},
	}})
	got := r.Apply(baseEvent())
	v, _ := got.Data["email"].(string)
	if len(v) != 16 {
		t.Fatalf("expected 16-char hex digest, got %q", v)
	}
}

func TestApply_PartialStrategy(t *testing.T) {
	r := New(Config{Rules: []Rule{
		{Table: "users", Columns: []string{"phone"}, Strategy: StrategyPartial},
	}})
	got := r.Apply(baseEvent())
	v, _ := got.Data["phone"].(string)
	if v[0] != '5' || v[len(v)-1] != '4' {
		t.Fatalf("unexpected partial redaction: %q", v)
	}
}

func TestApply_WildcardTable(t *testing.T) {
	r := New(Config{Rules: []Rule{
		{Table: "*", Columns: []string{"name"}, Strategy: StrategyBlank},
	}})
	ev := baseEvent()
	ev.Table = "orders"
	got := r.Apply(ev)
	if got.Data["name"] != "" {
		t.Fatalf("wildcard rule should match any table, got %v", got.Data["name"])
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	r := New(Config{Rules: []Rule{
		{Table: "users", Columns: []string{"email"}, Strategy: StrategyBlank},
	}})
	original := baseEvent()
	_ = r.Apply(original)
	if original.Data["email"] != "alice@example.com" {
		t.Fatal("original event must not be mutated")
	}
}
