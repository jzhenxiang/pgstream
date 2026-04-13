package matcher_test

import (
	"testing"

	"github.com/your-org/pgstream/internal/matcher"
)

func TestNew_EmptyPatterns_ReturnsError(t *testing.T) {
	_, err := matcher.New(nil)
	if err == nil {
		t.Fatal("expected error for nil patterns, got nil")
	}
	_, err = matcher.New([]string{})
	if err == nil {
		t.Fatal("expected error for empty patterns, got nil")
	}
}

func TestNew_BlankPattern_ReturnsError(t *testing.T) {
	_, err := matcher.New([]string{"public.*", ""})
	if err == nil {
		t.Fatal("expected error for blank pattern, got nil")
	}
}

func TestNew_Valid(t *testing.T) {
	m, err := matcher.New([]string{"public.*"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m == nil {
		t.Fatal("expected non-nil matcher")
	}
}

func TestMatch_ExactPattern(t *testing.T) {
	m, _ := matcher.New([]string{"public.users"})
	if !m.Match("public.users") {
		t.Error("expected match for exact pattern")
	}
	if m.Match("public.orders") {
		t.Error("expected no match for different table")
	}
}

func TestMatch_WildcardPattern(t *testing.T) {
	m, _ := matcher.New([]string{"public.*"})
	if !m.Match("public.users") {
		t.Error("expected wildcard match")
	}
	if !m.Match("public.orders") {
		t.Error("expected wildcard match for orders")
	}
	if m.Match("private.users") {
		t.Error("expected no match for different schema")
	}
}

func TestMatch_CaseInsensitive(t *testing.T) {
	m, _ := matcher.New([]string{"Public.Users"})
	if !m.Match("public.users") {
		t.Error("expected case-insensitive match")
	}
	if !m.Match("PUBLIC.USERS") {
		t.Error("expected case-insensitive match for uppercase input")
	}
}

func TestMatchAny_ReturnsTrueOnFirstMatch(t *testing.T) {
	m, _ := matcher.New([]string{"public.users"})
	values := []string{"public.orders", "public.users", "public.products"}
	if !m.MatchAny(values) {
		t.Error("expected MatchAny to return true")
	}
}

func TestMatchAny_ReturnsFalseWhenNoneMatch(t *testing.T) {
	m, _ := matcher.New([]string{"public.users"})
	values := []string{"public.orders", "public.products"}
	if m.MatchAny(values) {
		t.Error("expected MatchAny to return false")
	}
}

func TestPatterns_ReturnsCopy(t *testing.T) {
	m, _ := matcher.New([]string{"public.*", "private.*"})
	p := m.Patterns()
	if len(p) != 2 {
		t.Fatalf("expected 2 patterns, got %d", len(p))
	}
	p[0] = "mutated"
	if m.Patterns()[0] == "mutated" {
		t.Error("Patterns should return a copy, not a reference")
	}
}
