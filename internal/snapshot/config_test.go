package snapshot

import "testing"

func TestValidate_MissingDSN(t *testing.T) {
	cfg := &Config{Tables: []string{"users"}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing DSN")
	}
}

func TestValidate_MissingTables(t *testing.T) {
	cfg := &Config{DSN: "postgres://localhost/db"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing tables")
	}
}

func TestValidate_EmptyTableEntry(t *testing.T) {
	cfg := &Config{
		DSN:    "postgres://localhost/db",
		Tables: []string{"users", ""},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty table name")
	}
}

func TestValidate_NegativeBatchSize(t *testing.T) {
	cfg := &Config{
		DSN:       "postgres://localhost/db",
		Tables:    []string{"users"},
		BatchSize: -1,
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for negative BatchSize")
	}
}

func TestValidate_DefaultBatchSize(t *testing.T) {
	cfg := &Config{
		DSN:    "postgres://localhost/db",
		Tables: []string{"users"},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.BatchSize != 500 {
		t.Fatalf("expected default BatchSize 500, got %d", cfg.BatchSize)
	}
}

func TestValidate_Valid(t *testing.T) {
	cfg := &Config{
		DSN:       "postgres://localhost/db",
		Tables:    []string{"orders", "users"},
		BatchSize: 250,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
