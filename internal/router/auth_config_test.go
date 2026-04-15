package router

import (
	"testing"
)

func TestAuthConfig_Validate_Nil(t *testing.T) {
	var cfg *AuthConfig
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected nil error for nil config, got %v", err)
	}
}

func TestAuthConfig_Validate_NoTokens(t *testing.T) {
	cfg := &AuthConfig{Tokens: []string{}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty tokens, got nil")
	}
}

func TestAuthConfig_Validate_BlankTokensOnly(t *testing.T) {
	cfg := &AuthConfig{Tokens: []string{"  ", ""}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for blank-only tokens, got nil")
	}
}

func TestAuthConfig_Validate_ValidToken(t *testing.T) {
	cfg := &AuthConfig{Tokens: []string{"secret"}}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected nil error for valid config, got %v", err)
	}
}

func TestAuthConfig_Validate_MixedTokens(t *testing.T) {
	cfg := &AuthConfig{Tokens: []string{"", "valid-token"}}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected nil error when at least one valid token present, got %v", err)
	}
}
