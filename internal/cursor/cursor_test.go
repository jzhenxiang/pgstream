package cursor

import (
	"testing"
)

func TestNew_InitialisesAtStart(t *testing.T) {
	c := New(100)
	if c.Current() != 100 {
		t.Fatalf("expected current=100, got %d", c.Current())
	}
	if c.HighWaterMark() != 100 {
		t.Fatalf("expected hwm=100, got %d", c.HighWaterMark())
	}
}

func TestAdvance_MovesForward(t *testing.T) {
	c := New(0)
	moved := c.Advance(50)
	if !moved {
		t.Fatal("expected Advance to return true")
	}
	if c.Current() != 50 {
		t.Fatalf("expected current=50, got %d", c.Current())
	}
}

func TestAdvance_NoOpWhenEqual(t *testing.T) {
	c := New(10)
	moved := c.Advance(10)
	if moved {
		t.Fatal("expected Advance to return false for equal LSN")
	}
	if c.Current() != 10 {
		t.Fatalf("expected current unchanged at 10, got %d", c.Current())
	}
}

func TestAdvance_NoOpWhenBehind(t *testing.T) {
	c := New(20)
	moved := c.Advance(5)
	if moved {
		t.Fatal("expected Advance to return false for older LSN")
	}
	if c.Current() != 20 {
		t.Fatalf("expected current unchanged at 20, got %d", c.Current())
	}
}

func TestHighWaterMark_TracksMax(t *testing.T) {
	c := New(0)
	c.Advance(100)
	c.Reset(10)
	if c.HighWaterMark() != 100 {
		t.Fatalf("expected hwm=100 after reset, got %d", c.HighWaterMark())
	}
}

func TestReset_SetsCurrent(t *testing.T) {
	c := New(50)
	c.Advance(200)
	c.Reset(75)
	if c.Current() != 75 {
		t.Fatalf("expected current=75 after reset, got %d", c.Current())
	}
}

func TestLSN_String(t *testing.T) {
	var lsn LSN = 0x0000000100000001
	s := lsn.String()
	if s == "" {
		t.Fatal("expected non-empty LSN string")
	}
}
