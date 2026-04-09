package wal

import (
	"testing"
)

func TestNewDecoder(t *testing.T) {
	d := NewDecoder()
	if d == nil {
		t.Fatal("expected non-nil decoder")
	}
	if d.relations == nil {
		t.Fatal("expected relations map to be initialized")
	}
}

func TestDecodeEmptyMessage(t *testing.T) {
	d := NewDecoder()
	event, err := d.Decode(Message{Data: []byte{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event != nil {
		t.Fatalf("expected nil event for empty message, got %+v", event)
	}
}

func TestDecodeNonXLogData(t *testing.T) {
	d := NewDecoder()
	// 0x6b is the keepalive byte, not XLogData
	event, err := d.Decode(Message{Data: []byte{0x6b, 0x00}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event != nil {
		t.Fatalf("expected nil event for keepalive, got %+v", event)
	}
}

func TestMessageTypeConstants(t *testing.T) {
	tests := []struct {
		mt   MessageType
		want string
	}{
		{MessageTypeInsert, "INSERT"},
		{MessageTypeUpdate, "UPDATE"},
		{MessageTypeDelete, "DELETE"},
		{MessageTypeTruncate, "TRUNCATE"},
		{MessageTypeBegin, "BEGIN"},
		{MessageTypeCommit, "COMMIT"},
	}
	for _, tt := range tests {
		if string(tt.mt) != tt.want {
			t.Errorf("MessageType %q: got %q, want %q", tt.mt, string(tt.mt), tt.want)
		}
	}
}

func TestEventFields(t *testing.T) {
	e := Event{
		Type:    MessageTypeInsert,
		LSN:     "0/1234ABC",
		Schema:  "public",
		Table:   "users",
		Columns: map[string]any{"id": "1", "name": "alice"},
	}
	if e.Type != MessageTypeInsert {
		t.Errorf("expected INSERT, got %s", e.Type)
	}
	if e.Schema != "public" || e.Table != "users" {
		t.Errorf("unexpected schema/table: %s.%s", e.Schema, e.Table)
	}
	if len(e.Columns) != 2 {
		t.Errorf("expected 2 columns, got %d", len(e.Columns))
	}
}
