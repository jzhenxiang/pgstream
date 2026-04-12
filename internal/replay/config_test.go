package replay_test

import (
	"testing"

	"github.com/pgstream/pgstream/internal/replay"
)

func TestConfig_Validate_NilConfig(t *testing.T) {
	var c *replay.Config
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestConfig_Validate_EmptyOffsetFile(t *testing.T) {
	c := &replay.Config{StartLSN: "0/1"}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for empty offset_file")
	}
}

func TestConfig_Validate_WithStartLSN(t *testing.T) {
	c := &replay.Config{
		StartLSN:   "0/ABCDEF",
		OffsetFile: "/tmp/pgstream_offset",
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfig_Validate_OnlyOffsetFile(t *testing.T) {
	c := &replay.Config{OffsetFile: "/tmp/pgstream_offset"}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
