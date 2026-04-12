package partitioner

import "testing"

func TestConfig_Validate_NilConfig(t *testing.T) {
	var c *Config
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestConfig_Validate_ZeroPartitions(t *testing.T) {
	c := &Config{Strategy: StrategyTable, Partitions: 0}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for zero partitions")
	}
}

func TestConfig_Validate_UnknownStrategy(t *testing.T) {
	c := &Config{Strategy: "random", Partitions: 4}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for unknown strategy")
	}
}

func TestConfig_Validate_CustomMissingField(t *testing.T) {
	c := &Config{Strategy: StrategyCustom, Partitions: 4}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error when custom_field is missing")
	}
}

func TestConfig_Validate_ValidTable(t *testing.T) {
	c := &Config{Strategy: StrategyTable, Partitions: 8}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfig_Validate_ValidPK(t *testing.T) {
	c := &Config{Strategy: StrategyPK, Partitions: 4}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfig_Validate_ValidCustom(t *testing.T) {
	c := &Config{Strategy: StrategyCustom, Partitions: 4, CustomField: "tenant_id"}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDefaultConfig_Values(t *testing.T) {
	c := DefaultConfig()
	if c.Strategy != StrategyTable {
		t.Errorf("expected strategy %q, got %q", StrategyTable, c.Strategy)
	}
	if c.Partitions != 8 {
		t.Errorf("expected 8 partitions, got %d", c.Partitions)
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("default config invalid: %v", err)
	}
}
