package transform

import (
	"testing"
)

func TestNew_EmptyConfig(t *testing.T) {
	tr := New(Config{})
	if tr == nil {
		t.Fatal("expected non-nil Transformer")
	}
	if len(tr.redact) != 0 {
		t.Errorf("expected empty redact map, got %d entries", len(tr.redact))
	}
}

func TestApply_NilEvent(t *testing.T) {
	tr := New(Config{})
	_, err := tr.Apply("users", nil)
	if err == nil {
		t.Fatal("expected error for nil event")
	}
}

func TestApply_RedactsColumns(t *testing.T) {
	tr := New(Config{RedactColumns: []string{"password", "ssn"}})

	event := Event{"name": "alice", "password": "secret", "ssn": "123-45-6789"}
	out, err := tr.Apply("users", event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["password"] != "[REDACTED]" {
		t.Errorf("expected password redacted, got %v", out["password"])
	}
	if out["ssn"] != "[REDACTED]" {
		t.Errorf("expected ssn redacted, got %v", out["ssn"])
	}
	if out["name"] != "alice" {
		t.Errorf("expected name unchanged, got %v", out["name"])
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	tr := New(Config{RedactColumns: []string{"token"}})

	event := Event{"token": "abc123", "id": 1}
	_, err := tr.Apply("sessions", event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event["token"] != "abc123" {
		t.Error("original event was mutated")
	}
}

func TestApply_RenamesColumns(t *testing.T) {
	tr := New(Config{RenameColumns: map[string]string{"ts": "created_at"}})

	event := Event{"ts": "2024-01-01", "id": 42}
	out, err := tr.Apply("orders", event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := out["ts"]; ok {
		t.Error("original key 'ts' should have been removed")
	}
	if out["created_at"] != "2024-01-01" {
		t.Errorf("expected created_at=2024-01-01, got %v", out["created_at"])
	}
}

func TestApply_AddsMetadata(t *testing.T) {
	tr := New(Config{AddMetadata: true})

	out, err := tr.Apply("products", Event{"sku": "X1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["_pgstream_table"] != "products" {
		t.Errorf("expected _pgstream_table=products, got %v", out["_pgstream_table"])
	}
	if out["_pgstream_ts"] == nil {
		t.Error("expected _pgstream_ts to be set")
	}
}
