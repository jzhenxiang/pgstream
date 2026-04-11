package compressor

import (
	"bytes"
	"compress/gzip"
	"io"
	"strings"
	"testing"
)

func TestNew_DefaultAlgorithm(t *testing.T) {
	c, err := New(Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cAlgorithm() != None {
		t.Errorf("expected None, got %q", c.Algorithm())
	}
}

func TestNew_UnsupportedAlgorithm(t *testing.T) {
	_, err := New(Config{Algorithm: "zstd"})
	if err == nil {
		t.Fatal("expected error for unsupported algorithm")
	}
	if !strings.Contains(err.Error(), "unsupported algorithm") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCompress_None_ReturnsSameBytes(t *testing.T) {
	c, _ := New(Config{Algorithm: None})
	src := []byte("hello pgstream")
	out, err := c.Compress(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(src, out) {
		t.Errorf("expected passthrough, got different bytes")
	}
}

func TestCompress_Gzip_ProducesValidGzip(t *testing.T) {
	c, err := New(Config{Algorithm: Gzip})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := []byte("the quick brown fox jumps over the lazy dog")
	out, err := c.Compress(src)
	if err != nil {
		t.Fatalf("compress error: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("expected non-empty output")
	}
	// Decompress and verify round-trip.
	r, err := gzip.NewReader(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("gzip.NewReader: %v", err)
	}
	decompressed, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read decompressed: %v", err)
	}
	if !bytes.Equal(src, decompressed) {
		t.Errorf("round-trip mismatch: got %q, want %q", decompressed, src)
	}
}

func TestCompress_Gzip_EmptyInput(t *testing.T) {
	c, _ := New(Config{Algorithm: Gzip})
	out, err := c.Compress([]byte{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// A gzip stream for empty input is still valid gzip.
	r, err := gzip.NewReader(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("gzip.NewReader on empty: %v", err)
	}
	decompressed, _ := io.ReadAll(r)
	if len(decompressed) != 0 {
		t.Errorf("expected empty decompressed, got %d bytes", len(decompressed))
	}
}

func TestAlgorithm_ReturnsConfigured(t *testing.T) {
	for _, alg := range []Algorithm{None, Gzip} {
		c, err := New(Config{Algorithm: alg})
		if err != nil {
			t.Fatalf("New(%q): %v", alg, err)
		}
		if c.Algorithm() != alg {
			t.Errorf("Algorithm(): got %q, want %q", c.Algorithm(), alg)
		}
	}
}
