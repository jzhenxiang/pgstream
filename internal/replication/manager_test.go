package replication

import (
	"context"
	"log"
	"os"
	"testing"
)

func testLogger() *log.Logger {
	return log.New(os.Stdout, "test: ", 0)
}

func TestNewManager_NilSlot(t *testing.T) {
	_, err := NewManager(nil, testLogger())
	if err == nil {
		t.Fatal("expected error for nil slot")
	}
}

func TestNewManager_DefaultLogger(t *testing.T) {
	slot := &Slot{cfg: SlotConfig{SlotName: "pgstream_slot"}}
	m, err := NewManager(slot, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.logger == nil {
		t.Error("expected default logger to be set")
	}
}

func TestNewManager_Valid(t *testing.T) {
	slot := &Slot{cfg: SlotConfig{SlotName: "pgstream_slot"}}
	m, err := NewManager(slot, testLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.SlotName() != "pgstream_slot" {
		t.Errorf("expected slot name pgstream_slot, got %s", m.SlotName())
	}
}

func TestEnsureSlot_EmptyName(t *testing.T) {
	slot := &Slot{cfg: SlotConfig{SlotName: ""}}
	m, _ := NewManager(slot, testLogger())
	err := m.EnsureSlot(context.Background())
	if err == nil {
		t.Fatal("expected error for empty slot name")
	}
}

func TestEnsureSlot_ValidName(t *testing.T) {
	slot := &Slot{cfg: SlotConfig{SlotName: "pgstream_slot"}}
	m, _ := NewManager(slot, testLogger())
	err := m.EnsureSlot(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSlotName_ReturnsCorrectName(t *testing.T) {
	slot := &Slot{cfg: SlotConfig{SlotName: "my_replication_slot"}}
	m, _ := NewManager(slot, testLogger())
	if m.SlotName() != "my_replication_slot" {
		t.Errorf("expected my_replication_slot, got %s", m.SlotName())
	}
}
