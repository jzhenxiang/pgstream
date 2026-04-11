package masker

import (
	"fmt"
	"strings"
	"testing"

	"github.com/your-org/pgstream/internal/wal"
)

func baseEvent() *wal.Event {
	return &wal.Event{
		Table: "users",
		Fields: map[string]interface{}{
			"id":    1,
			"email": "alice@example.com",
			"name":  "Alice",
		},
	}
}

func TestNew_EmptyConfig(t *testing.T) {
	m := New(Config{})
	if m == nil {
		t.Fatal("expected non-nil Masker")
	}
	if len(m.rules) != 0 {
		t.Errorf("expected 0 rules, got %d", len(m.rules))
	}
}

func TestApply_NilEvent_ReturnsNil(t *testing.T) {
	m := New(Config{Rules: []Rule{{Table: "users", Column: "email", Strategy: StrategyRedact}}})
	if got := m.Apply(nil); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestApply_NoMatchingRules_ReturnsSamePointer(t *testing.T) {
	m := New(Config{Rules: []Rule{{Table: "orders", Column: "total", Strategy: StrategyRedact}}})
	ev := baseEvent()
	got := m.Apply(ev)
	if got != ev {
		t.Error("expected same pointer when no rules match")
	}
}

func TestApply_Redact(t *testing.T) {
	m := New(Config{Rules: []Rule{{Table: "users", Column: "email", Strategy: StrategyRedact}}})
	got := m.Apply(baseEvent())
	if got.Fields["email"] != defaultRedactValue {
		t.Errorf("expected %q, got %v", defaultRedactValue, got.Fields["email"])
	}
	if got.Fields["name"] != "Alice" {
		t.Error("non-masked field should be unchanged")
	}
}

func TestApply_CustomRedactValue(t *testing.T) {
	m := New(Config{Rules: []Rule{{Table: "users", Column: "email", Strategy: StrategyRedact, RedactValue: "***"}}})
	got := m.Apply(baseEvent())
	if got.Fields["email"] != "***" {
		t.Errorf("expected ***, got %v", got.Fields["email"])
	}
}

func TestApply_Hash(t *testing.T) {
	m := New(Config{Rules: []Rule{{Table: "users", Column: "email", Strategy: StrategyHash}}})
	got := m.Apply(baseEvent())
	hashed, ok := got.Fields["email"].(string)
	if !ok {
		t.Fatal("expected string")
	}
	if len(hashed) != 64 {
		t.Errorf("expected 64-char hex digest, got len=%d", len(hashed))
	}
}

func TestApply_Partial(t *testing.T) {
	m := New(Config{Rules: []Rule{{Table: "users", Column: "email", Strategy: StrategyPartial, PartialKeep: 3}}})
	got := m.Apply(baseEvent())
	v := fmt.Sprintf("%v", got.Fields["email"])
	if !strings.HasPrefix(v, "ali") {
		t.Errorf("expected prefix 'ali', got %q", v)
	}
	if strings.Contains(v[3:], "@") {
		t.Errorf("expected remainder masked, got %q", v)
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	m := New(Config{Rules: []Rule{{Table: "users", Column: "email", Strategy: StrategyRedact}}})
	orig := baseEvent()
	origEmail := orig.Fields["email"]
	m.Apply(orig)
	if orig.Fields["email"] != origEmail {
		t.Error("original event should not be mutated")
	}
}
