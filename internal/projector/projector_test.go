package projector_test

import (
	"testing"

	"github.com/your-org/pgstream/internal/projector"
	"github.com/your-org/pgstream/internal/wal"
)

func baseEvent() *wal.Event {
	return &wal.Event{
		Table: "public.orders",
		Data: map[string]any{
			"id":         1,
			"status":     "pending",
			"total":      99.9,
			"created_at": "2024-01-01",
		},
	}
}

func TestNew_EmptyConfig(t *testing.T) {
	p, err := projector.New(projector.Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil projector")
	}
}

func TestNew_BlankTableName_ReturnsError(t *testing.T) {
	_, err := projector.New(projector.Config{
		Rules: map[string][]string{"": {"id"}},
	})
	if err == nil {
		t.Fatal("expected error for blank table name")
	}
}

func TestNew_EmptyColumnList_ReturnsError(t *testing.T) {
	_, err := projector.New(projector.Config{
		Rules: map[string][]string{"orders": {}},
	})
	if err == nil {
		t.Fatal("expected error for empty column list")
	}
}

func TestApply_NilEvent_ReturnsNil(t *testing.T) {
	p, _ := projector.New(projector.Config{})
	if got := p.Apply(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestApply_NoMatchingRule_ReturnsSamePointer(t *testing.T) {
	p, _ := projector.New(projector.Config{})
	ev := baseEvent()
	got := p.Apply(ev)
	if got != ev {
		t.Fatal("expected same pointer when no rule matches")
	}
}

func TestApply_RetainsOnlyAllowedColumns(t *testing.T) {
	p, _ := projector.New(projector.Config{
		Rules: map[string][]string{
			"public.orders": {"id", "status"},
		},
	})
	out := p.Apply(baseEvent())
	if _, ok := out.Data["id"]; !ok {
		t.Error("expected 'id' to be retained")
	}
	if _, ok := out.Data["status"]; !ok {
		t.Error("expected 'status' to be retained")
	}
	if _, ok := out.Data["total"]; ok {
		t.Error("expected 'total' to be removed")
	}
	if _, ok := out.Data["created_at"]; ok {
		t.Error("expected 'created_at' to be removed")
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	p, _ := projector.New(projector.Config{
		Rules: map[string][]string{
			"public.orders": {"id"},
		},
	})
	ev := baseEvent()
	origLen := len(ev.Data)
	p.Apply(ev)
	if len(ev.Data) != origLen {
		t.Fatal("Apply must not mutate the original event")
	}
}
