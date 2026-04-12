package enricher

import (
	"strings"
	"testing"
	"time"

	"github.com/pgstream/pgstream/internal/wal"
)

func baseEvent() *wal.Event {
	return &wal.Event{
		Table:    "orders",
		Metadata: map[string]string{"existing": "value"},
	}
}

func TestNew_NoHostname(t *testing.T) {
	e, err := New(Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e == nil {
		t.Fatal("expected non-nil enricher")
	}
}

func TestApply_NilEvent_ReturnsNil(t *testing.T) {
	e, _ := New(Config{})
	if got := e.Apply(nil); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestApply_StaticFields(t *testing.T) {
	e, err := New(Config{
		StaticFields: map[string]string{"env": "prod", "region": "us-east-1"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := e.Apply(baseEvent())
	if out.Metadata["env"] != "prod" {
		t.Errorf("expected env=prod, got %q", out.Metadata["env"])
	}
	if out.Metadata["region"] != "us-east-1" {
		t.Errorf("expected region=us-east-1, got %q", out.Metadata["region"])
	}
	if out.Metadata["existing"] != "value" {
		t.Error("existing metadata key should be preserved")
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	e, _ := New(Config{StaticFields: map[string]string{"k": "v"}})
	orig := baseEvent()
	e.Apply(orig)
	if _, ok := orig.Metadata["k"]; ok {
		t.Error("original event metadata was mutated")
	}
}

func TestApply_AddTimestamp(t *testing.T) {
	e, _ := New(Config{AddTimestamp: true})
	fixed := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	e.now = func() time.Time { return fixed }

	out := e.Apply(baseEvent())
	got := out.Metadata["_enriched_at"]
	if !strings.HasPrefix(got, "2024-01-15") {
		t.Errorf("unexpected _enriched_at value: %q", got)
	}
}

func TestApply_AddHostname(t *testing.T) {
	e, _ := New(Config{})
	e.hostname = "test-host"
	e.cfg.AddHostname = true

	out := e.Apply(baseEvent())
	if out.Metadata["_hostname"] != "test-host" {
		t.Errorf("expected _hostname=test-host, got %q", out.Metadata["_hostname"])
	}
}

func TestApply_NoHostname_WhenDisabled(t *testing.T) {
	e, _ := New(Config{AddHostname: false})
	out := e.Apply(baseEvent())
	if _, ok := out.Metadata["_hostname"]; ok {
		t.Error("_hostname should not be set when AddHostname is false")
	}
}
