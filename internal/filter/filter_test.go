package filter_test

import (
	"testing"

	"github.com/your-org/pgstream/internal/filter"
)

func TestNew_EmptyConfig(t *testing.T) {
	f := filter.New(filter.Config{})
	if !f.IsEmpty() {
		t.Fatal("expected filter to be empty")
	}
}

func TestAllow_NoRules_AllowsAll(t *testing.T) {
	f := filter.New(filter.Config{})
	if !f.Allow("public", "users") {
		t.Error("expected all tables to be allowed when no rules set")
	}
}

func TestAllow_AllowList(t *testing.T) {
	f := filter.New(filter.Config{
		AllowTables: []string{"public.orders", "public.products"},
	})

	if !f.Allow("public", "orders") {
		t.Error("expected public.orders to be allowed")
	}
	if f.Allow("public", "users") {
		t.Error("expected public.users to be denied (not in allow list)")
	}
}

func TestAllow_DenyList(t *testing.T) {
	f := filter.New(filter.Config{
		DenyTables: []string{"public.audit_log"},
	})

	if f.Allow("public", "audit_log") {
		t.Error("expected public.audit_log to be denied")
	}
	if !f.Allow("public", "users") {
		t.Error("expected public.users to be allowed")
	}
}

func TestAllow_DenyOverridesAllow(t *testing.T) {
	f := filter.New(filter.Config{
		AllowTables: []string{"public.orders"},
		DenyTables:  []string{"public.orders"},
	})

	if f.Allow("public", "orders") {
		t.Error("expected deny list to override allow list")
	}
}

func TestAllow_CaseInsensitive(t *testing.T) {
	f := filter.New(filter.Config{
		AllowTables: []string{"Public.Orders"},
	})

	if !f.Allow("public", "orders") {
		t.Error("expected case-insensitive match for public.orders")
	}
}

func TestIsEmpty_WithRules(t *testing.T) {
	f := filter.New(filter.Config{
		AllowTables: []string{"public.events"},
	})
	if f.IsEmpty() {
		t.Error("expected filter to not be empty")
	}
}
