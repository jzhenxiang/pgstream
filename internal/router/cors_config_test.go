package router

import "testing"

func TestCORSConfig_Validate_Nil(t *testing.T) {
	var cfg *CORSConfig
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected nil error for nil config, got %v", err)
	}
}

func TestCORSConfig_Validate_Valid(t *testing.T) {
	cfg := &CORSConfig{
		AllowedOrigins: []string{"https://example.com"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         300,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCORSConfig_Validate_BlankOrigin_ReturnsError(t *testing.T) {
	cfg := &CORSConfig{AllowedOrigins: []string{"https://ok.com", ""}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for blank origin")
	}
}

func TestCORSConfig_Validate_BlankMethod_ReturnsError(t *testing.T) {
	cfg := &CORSConfig{
		AllowedOrigins: []string{"https://ok.com"},
		AllowedMethods: []string{"GET", ""},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for blank method")
	}
}

func TestCORSConfig_Validate_BlankHeader_ReturnsError(t *testing.T) {
	cfg := &CORSConfig{
		AllowedOrigins: []string{"https://ok.com"},
		AllowedHeaders: []string{""},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for blank header")
	}
}

func TestCORSConfig_Validate_NegativeMaxAge_ReturnsError(t *testing.T) {
	cfg := &CORSConfig{
		AllowedOrigins: []string{"https://ok.com"},
		MaxAge:         -1,
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for negative MaxAge")
	}
}

func TestCORSConfig_Validate_WildcardOrigin_Valid(t *testing.T) {
	cfg := &CORSConfig{AllowedOrigins: []string{"*"}}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no error for wildcard, got %v", err)
	}
}
