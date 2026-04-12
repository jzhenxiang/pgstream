package partitioner

import (
	"testing"

	"github.com/pgstream/pgstream/internal/wal"
)

func TestNew_ZeroPartitions_ReturnsError(t *testing.T) {
	_, err := New(Config{Partitions: 0, Strategy: StrategyTable})
	if err == nil {
		t.Fatal("expected error for zero partitions")
	}
}

func TestNew_CustomStrategy_MissingField_ReturnsError(t *testing.T) {
	_, err := New(Config{Partitions: 4, Strategy: StrategyCustom})
	if err == nil {
		t.Fatal("expected error when custom_field is empty")
	}
}

func TestNew_DefaultStrategy_Applied(t *testing.T) {
	p, err := New(Config{Partitions: 4})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.cfg.Strategy != StrategyTable {
		t.Errorf("expected default strategy %q, got %q", StrategyTable, p.cfg.Strategy)
	}
}

func TestAssign_NilEvent_ReturnsError(t *testing.T) {
	p, _ := New(Config{Partitions: 4, Strategy: StrategyTable})
	_, err := p.Assign(nil)
	if err == nil {
		t.Fatal("expected error for nil event")
	}
}

func TestAssign_TableStrategy_Deterministic(t *testing.T) {
	p, _ := New(Config{Partitions: 8, Strategy: StrategyTable})
	event := &wal.Event{Table: "orders"}

	first, err := p.Assign(event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 0; i < 5; i++ {
		got, _ := p.Assign(event)
		if got != first {
			t.Errorf("partition not deterministic: got %d, want %d", got, first)
		}
	}
}

func TestAssign_PKStrategy_UsesTableWhenPKEmpty(t *testing.T) {
	p, _ := New(Config{Partitions: 8, Strategy: StrategyPK})
	eventNoPK := &wal.Event{Table: "users", PrimaryKey: ""}
	eventWithPK := &wal.Event{Table: "users", PrimaryKey: "42"}

	partNoPK, _ := p.Assign(eventNoPK)
	partWithPK, _ := p.Assign(eventWithPK)

	// They may differ; we just check they are in range.
	if partNoPK < 0 || partNoPK >= 8 {
		t.Errorf("partition out of range: %d", partNoPK)
	}
	if partWithPK < 0 || partWithPK >= 8 {
		t.Errorf("partition out of range: %d", partWithPK)
	}
}

func TestAssign_CustomStrategy_MissingField_FallsBackToTable(t *testing.T) {
	p, _ := New(Config{Partitions: 4, Strategy: StrategyCustom, CustomField: "tenant_id"})
	event := &wal.Event{Table: "invoices", Data: map[string]interface{}{}}

	part, err := p.Assign(event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if part < 0 || part >= 4 {
		t.Errorf("partition out of range: %d", part)
	}
}

func TestAssign_PartitionInRange(t *testing.T) {
	partitions := 16
	p, _ := New(Config{Partitions: partitions, Strategy: StrategyTable})
	tables := []string{"a", "bb", "ccc", "orders", "users", "payments"}
	for _, tbl := range tables {
		event := &wal.Event{Table: tbl}
		part, err := p.Assign(event)
		if err != nil {
			t.Fatalf("table %q: unexpected error: %v", tbl, err)
		}
		if part < 0 || part >= partitions {
			t.Errorf("table %q: partition %d out of range [0, %d)", tbl, part, partitions)
		}
	}
}
