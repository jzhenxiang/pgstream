package selector

import (
	"testing"

	"pgstream/internal/wal"
)

func baseEvent() *wal.Event {
	return &wal.Event{
		Schema: "public",
		Table:  "users",
		Columns: []wal.Column{
			{Name: "id", Value: 1},
			{Name: "email", Value: "a@b.com"},
			{Name: "password", Value: "secret"},
		},
	}
}

func TestNew_EmptyConfig(t *testing.T) {
	s, err := New(Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil selector")
	}
}

func TestNew_EmptyTableName_ReturnsError(t *testing.T) {
	_, err := New(Config{Rules: map[string][]string{"": {"id"}}})
	if err == nil {
		t.Fatal("expected error for empty table name")
	}
}

func TestApply_NilEvent_ReturnsNil(t *testing.T) {
	s, _ := New(Config{})
	if got := s.Apply(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestApply_NoRules_ReturnsSamePointer(t *testing.T) {
	s, _ := New(Config{})
	ev := baseEvent()
	if got := s.Apply(ev); got != ev {
		t.Fatal("expected same pointer when no rules configured")
	}
}

func TestApply_RetainsAllowedColumns(t *testing.T) {
	s, err := New(Config{Rules: map[string][]string{
		"users": {"id", "email"},
	}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := s.Apply(baseEvent())
	if len(got.Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(got.Columns))
	}
	for _, col := range got.Columns {
		if col.Name == "password" {
			t.Fatal("password column should have been removed")
		}
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	s, _ := New(Config{Rules: map[string][]string{"users": {"id"}}})
	ev := baseEvent()
	orig := len(ev.Columns)
	s.Apply(ev)
	if len(ev.Columns) != orig {
		t.Fatal("original event columns were mutated")
	}
}

func TestApply_SchemaQualifiedTableName(t *testing.T) {
	s, err := New(Config{Rules: map[string][]string{
		"public.users": {"id"},
	}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := s.Apply(baseEvent())
	if len(got.Columns) != 1 || got.Columns[0].Name != "id" {
		t.Fatalf("expected only id column, got %+v", got.Columns)
	}
}

func TestApply_UnknownTable_PassesThrough(t *testing.T) {
	s, _ := New(Config{Rules: map[string][]string{"orders": {"id"}}})
	ev := baseEvent()
	got := s.Apply(ev)
	if got != ev {
		t.Fatal("expected same pointer for unmatched table")
	}
}
