package router

import (
	"testing"
)

func TestCORSConfig_Validate_Nil(t *testing.T) {
	var c *CORSConfig
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestCORSConfig_Validate_Valid(t *testing.T) {
	c := &CORSConfig{
		AllowedOrigins: []string{"https://example.com"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type"},
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCORSConfig_Validate_BlankOrigin_ReturnsError(t *testing.T) {
	c := &CORSConfig{
		AllowedOrigins: []string{""},
	}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for blank origin")
	}
}

func TestCORSConfig_Validate_BlankMethod_ReturnsError(t *testing.T) {
	c := &CORSConfig{
		AllowedOrigins: []string{"https://example.com"},
		AllowedMethods: []string{""},
	}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for blank method")
	}
}

func TestCORSConfig_Validate_BlankHeader_ReturnsError(t *testing.T) {
	c := &CORSConfig{
		AllowedOrigins: []string{"https://example.com"},
		AllowedHeaders: []string{""},
	}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for blank header")
	}
}

func TestCORSConfig_Validate_WildcardOrigin_Valid(t *testing.T) {
	c := &CORSConfig{
		AllowedOrigins: []string{"*"},
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCORSConfig_Validate_EmptyOrigins_Valid(t *testing.T) {
	c := &CORSConfig{}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error for empty config: %v", err)
	}
}
