package lsn

import (
	"testing"
)

func TestParse_Valid(t *testing.T) {
	cases := []struct {
		input string
		want  LSN
	}{
		{"0/0", LSN(0)},
		{"0/1", LSN(1)},
		{"1/0", LSN(1) << 32},
		{"0/1A2B3C4D", LSN(0x1A2B3C4D)},
		{"A/BCDEF012", LSN(0xA_BCDEF012)},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := Parse(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("Parse(%q) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}

func TestParse_Invalid(t *testing.T) {
	cases := []string{
		"",
		"0",
		"ZZ/00",
		"00/ZZ",
		"1/2/3",
	}

	for _, input := range cases {
		t.Run(input, func(t *testing.T) {
			_, err := Parse(input)
			if err == nil {
				t.Errorf("Parse(%q): expected error, got nil", input)
			}
		})
	}
}

func TestLSN_String(t *testing.T) {
	l := MustParse("A/BCDEF012")
	got := l.String()
	if got != "A/BCDEF012" {
		t.Errorf("String() = %q, want %q", got, "A/BCDEF012")
	}
}

func TestLSN_AfterBefore(t *testing.T) {
	a := MustParse("0/1")
	b := MustParse("0/2")

	if !b.After(a) {
		t.Error("expected b.After(a)")
	}
	if !a.Before(b) {
		t.Error("expected a.Before(b)")
	}
	if a.After(b) {
		t.Error("did not expect a.After(b)")
	}
}

func TestLSN_IsZero(t *testing.T) {
	if !Zero.IsZero() {
		t.Error("expected Zero.IsZero()")
	}
	if MustParse("0/1").IsZero() {
		t.Error("did not expect non-zero LSN to be zero")
	}
}

func TestMax(t *testing.T) {
	a := MustParse("0/10")
	b := MustParse("0/20")

	if Max(a, b) != b {
		t.Errorf("Max(a,b): expected b")
	}
	if Max(b, a) != b {
		t.Errorf("Max(b,a): expected b")
	}
	if Max(a, a) != a {
		t.Errorf("Max(a,a): expected a")
	}
}

func TestMustParse_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic, got none")
		}
	}()
	MustParse("invalid")
}
