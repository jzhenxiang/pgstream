package encoder

import (
	"encoding/json"
	"testing"

	"github.com/pgstream/pgstream/internal/wal"
)

func sampleEvent() *wal.Event {
	return &wal.Event{
		Table:  "users",
		Action: "INSERT",
		Data:   map[string]any{"id": 1, "name": "alice"},
	}
}

func TestNew_ReturnsEncoder(t *testing.T) {
	enc := New()
	if enc == nil {
		t.Fatal("expected non-nil encoder")
	}
	if enc.version != 1 {
		t.Fatalf("expected version 1, got %d", enc.version)
	}
}

func TestEncode_NilEvent_ReturnsError(t *testing.T) {
	enc := New()
	_, err := enc.Encode(nil)
	if err == nil {
		t.Fatal("expected error for nil event")
	}
}

func TestEncode_ValidEvent_ReturnsJSON(t *testing.T) {
	enc := New()
	b, err := enc.Encode(sampleEvent())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(b, &env); err != nil {
		t.Fatalf("could not unmarshal envelope: %v", err)
	}

	if env.Version != 1 {
		t.Errorf("expected version 1, got %d", env.Version)
	}
	if env.Event == nil {
		t.Error("expected non-nil event in envelope")
	}
	if env.Event.Table != "users" {
		t.Errorf("expected table 'users', got %q", env.Event.Table)
	}
	if env.Meta != nil {
		t.Error("expected nil meta for plain Encode")
	}
}

func TestEncodeWithMeta_NilEvent_ReturnsError(t *testing.T) {
	enc := New()
	_, err := enc.EncodeWithMeta(nil, map[string]any{"k": "v"})
	if err == nil {
		t.Fatal("expected error for nil event")
	}
}

func TestEncodeWithMeta_SetsMetadata(t *testing.T) {
	enc := New()
	meta := map[string]any{"source": "pgstream", "slot": "test_slot"}
	b, err := enc.EncodeWithMeta(sampleEvent(), meta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(b, &env); err != nil {
		t.Fatalf("could not unmarshal envelope: %v", err)
	}

	if env.Meta == nil {
		t.Fatal("expected non-nil meta")
	}
	if env.Meta["source"] != "pgstream" {
		t.Errorf("unexpected meta source: %v", env.Meta["source"])
	}
	if env.Meta["slot"] != "test_slot" {
		t.Errorf("unexpected meta slot: %v", env.Meta["slot"])
	}
}

func TestEncode_TimestampIsSet(t *testing.T) {
	enc := New()
	b, err := enc.Encode(sampleEvent())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(b, &env); err != nil {
		t.Fatalf("could not unmarshal envelope: %v", err)
	}

	if env.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}
