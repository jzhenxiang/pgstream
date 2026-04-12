package validator_test

import (
	"errors"
	"testing"

	"github.com/pgstream/pgstream/internal/validator"
	"github.com/pgstream/pgstream/internal/wal"
)

func TestNew_EmptyConfig(t *testing.T) {
	v := validator.New(validator.Config{})
	if v.RuleCount() != 0 {
		t.Fatalf("expected 0 rules, got %d", v.RuleCount())
	}
}

func TestNew_WithRules(t *testing.T) {
	v := validator.New(validator.Config{
		Rules: []validator.Rule{
			{Table: "public.orders", RequiredColumns: []string{"id"}},
			{Table: "public.users", RequiredColumns: []string{"email"}},
		},
	})
	if v.RuleCount() != 2 {
		t.Fatalf("expected 2 rules, got %d", v.RuleCount())
	}
}

func TestValidate_NilEvent(t *testing.T) {
	v := validator.New(validator.Config{})
	if err := v.Validate(nil); err != nil {
		t.Fatalf("unexpected error for nil event: %v", err)
	}
}

func TestValidate_NoMatchingRule_Passes(t *testing.T) {
	v := validator.New(validator.Config{
		Rules: []validator.Rule{
			{Table: "public.orders", RequiredColumns: []string{"id"}},
		},
	})
	event := &wal.Event{Table: "public.products", Data: map[string]interface{}{}}
	if err := v.Validate(event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_AllColumnsPresent_Passes(t *testing.T) {
	v := validator.New(validator.Config{
		Rules: []validator.Rule{
			{Table: "public.orders", RequiredColumns: []string{"id", "status"}},
		},
	})
	event := &wal.Event{
		Table: "public.orders",
		Data:  map[string]interface{}{"id": 1, "status": "pending"},
	}
	if err := v.Validate(event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_MissingColumn_Fails(t *testing.T) {
	v := validator.New(validator.Config{
		Rules: []validator.Rule{
			{Table: "public.orders", RequiredColumns: []string{"id", "status"}},
		},
	})
	event := &wal.Event{
		Table: "public.orders",
		Data:  map[string]interface{}{"id": 1},
	}
	err := v.Validate(event)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !errors.Is(err, validator.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestValidate_NilColumnValue_Fails(t *testing.T) {
	v := validator.New(validator.Config{
		Rules: []validator.Rule{
			{Table: "public.orders", RequiredColumns: []string{"id"}},
		},
	})
	event := &wal.Event{
		Table: "public.orders",
		Data:  map[string]interface{}{"id": nil},
	}
	err := v.Validate(event)
	if !errors.Is(err, validator.ErrValidation) {
		t.Fatalf("expected ErrValidation for nil value, got %v", err)
	}
}

func TestValidate_EmptyRequiredColumns_Passes(t *testing.T) {
	v := validator.New(validator.Config{
		Rules: []validator.Rule{
			{Table: "public.logs", RequiredColumns: []string{}},
		},
	})
	event := &wal.Event{Table: "public.logs", Data: map[string]interface{}{}}
	if err := v.Validate(event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
