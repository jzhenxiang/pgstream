package rewriter_test

import (
	"testing"

	"pgstream/internal/rewriter"
	"pgstream/internal/wal"
)

func baseEvent() *wal.Event {
	return &wal.Event{
		Table:  "users",
		Fields: map[string]interface{}{"status": "active", "role": "admin"},
	}
}

func TestNew_EmptyConfig(t *testing.T) {
	r := rewriter.New(rewriter.Config{})
	if r == nil {
		t.Fatal("expected non-nil rewriter")
	}
}

func TestApply_NilEvent_ReturnsNil(t *testing.T) {
	r := rewriter.New(rewriter.Config{})
	if got := r.Apply(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestApply_NoRules_ReturnsSamePointer(t *testing.T) {
	r := rewriter.New(rewriter.Config{})
	ev := baseEvent()
	if got := r.Apply(ev); got != ev {
		t.Fatal("expected same pointer when no rules configured")
	}
}

func TestApply_RewritesMatchingField(t *testing.T) {
	r := rewriter.New(rewriter.Config{
		Rules: []rewriter.Rule{
			{Table: "users", Column: "status", Mapping: map[string]string{"active": "enabled"}},
		},
	})
	ev := baseEvent()
	got := r.Apply(ev)
	if got == ev {
		t.Fatal("expected a new copy")
	}
	if got.Fields["status"] != "enabled" {
		t.Fatalf("expected 'enabled', got %v", got.Fields["status"])
	}
	// original must be untouched
	if ev.Fields["status"] != "active" {
		t.Fatal("original event was mutated")
	}
}

func TestApply_TableMismatch_NoRewrite(t *testing.T) {
	r := rewriter.New(rewriter.Config{
		Rules: []rewriter.Rule{
			{Table: "orders", Column: "status", Mapping: map[string]string{"active": "enabled"}},
		},
	})
	ev := baseEvent()
	got := r.Apply(ev)
	if got != ev {
		t.Fatal("expected same pointer when table does not match")
	}
}

func TestApply_EmptyTableRule_MatchesAnyTable(t *testing.T) {
	r := rewriter.New(rewriter.Config{
		Rules: []rewriter.Rule{
			{Table: "", Column: "role", Mapping: map[string]string{"admin": "superuser"}},
		},
	})
	ev := baseEvent()
	got := r.Apply(ev)
	if got.Fields["role"] != "superuser" {
		t.Fatalf("expected 'superuser', got %v", got.Fields["role"])
	}
}

func TestApply_ValueNotInMapping_NoChange(t *testing.T) {
	r := rewriter.New(rewriter.Config{
		Rules: []rewriter.Rule{
			{Table: "users", Column: "status", Mapping: map[string]string{"inactive": "disabled"}},
		},
	})
	ev := baseEvent()
	got := r.Apply(ev)
	if got != ev {
		t.Fatal("expected same pointer when value not in mapping")
	}
}
