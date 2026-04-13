package typecast

import (
	"testing"
	"time"
)

func TestNew_DefaultTimeFmt(t *testing.T) {
	c := New("")
	if c.timeFmt != time.RFC3339 {
		t.Fatalf("expected RFC3339, got %q", c.timeFmt)
	}
}

func TestNew_CustomTimeFmt(t *testing.T) {
	c := New("2006-01-02")
	if c.timeFmt != "2006-01-02" {
		t.Fatalf("expected custom fmt, got %q", c.timeFmt)
	}
}

func TestToString(t *testing.T) {
	c := New("")
	cases := []struct {
		input interface{}
		want  string
	}{
		{nil, ""},
		{"hello", "hello"},
		{[]byte("bytes"), "bytes"},
		{int64(42), "42"},
		{float64(3.14), "3.14"},
		{true, "true"},
	}
	for _, tc := range cases {
		got, err := c.ToString(tc.input)
		if err != nil {
			t.Fatalf("ToString(%v): unexpected error: %v", tc.input, err)
		}
		if got != tc.want {
			t.Fatalf("ToString(%v): got %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestToString_Time(t *testing.T) {
	c := New("2006-01-02")
	tm := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	got, err := c.ToString(tm)
	if err != nil {
		t.Fatal(err)
	}
	if got != "2024-06-01" {
		t.Fatalf("got %q", got)
	}
}

func TestToInt64(t *testing.T) {
	c := New("")
	cases := []struct {
		input interface{}
		want  int64
	}{
		{nil, 0},
		{int64(7), 7},
		{float64(9.9), 9},
		{"123", 123},
		{true, 1},
		{false, 0},
	}
	for _, tc := range cases {
		got, err := c.ToInt64(tc.input)
		if err != nil {
			t.Fatalf("ToInt64(%v): %v", tc.input, err)
		}
		if got != tc.want {
			t.Fatalf("ToInt64(%v): got %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestToInt64_InvalidString(t *testing.T) {
	c := New("")
	_, err := c.ToInt64("notanumber")
	if err == nil {
		t.Fatal("expected error for invalid string")
	}
}

func TestToInt64_UnsupportedType(t *testing.T) {
	c := New("")
	_, err := c.ToInt64(struct{}{})
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestToBool(t *testing.T) {
	c := New("")
	cases := []struct {
		input interface{}
		want  bool
	}{
		{nil, false},
		{true, true},
		{false, false},
		{int64(1), true},
		{int64(0), false},
		{float64(1.0), true},
		{"true", true},
		{"false", false},
	}
	for _, tc := range cases {
		got, err := c.ToBool(tc.input)
		if err != nil {
			t.Fatalf("ToBool(%v): %v", tc.input, err)
		}
		if got != tc.want {
			t.Fatalf("ToBool(%v): got %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestToBool_UnsupportedType(t *testing.T) {
	c := New("")
	_, err := c.ToBool(struct{}{})
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}
