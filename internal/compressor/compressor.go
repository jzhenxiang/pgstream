// Package compressor provides optional payload compression for sink events
// before they are dispatched. Supported algorithms: gzip, zstd, none.
package compressor

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

// Algorithm represents a supported compression algorithm.
type Algorithm string

const (
	None Algorithm = "none"
	Gzip Algorithm = "gzip"
)

// Config holds compressor configuration.
type Config struct {
	// Algorithm selects the compression algorithm. Defaults to "none".
	Algorithm Algorithm
	// Level is the compression level (algorithm-specific). 0 means default.
	Level int
}

// Compressor compresses byte payloads using the configured algorithm.
type Compressor struct {
	cfg Config
}

// New returns a new Compressor for the given config.
// An error is returned if the algorithm is unsupported.
func New(cfg Config) (*Compressor, error) {
	if cfg.Algorithm == "" {
		cfg.Algorithm = None
	}
	switch cfg.Algorithm {
	case None, Gzip:
		// valid
	default:
		return nil, fmt.Errorf("compressor: unsupported algorithm %q", cfg.Algorithm)
	}
	return &Compressor{cfg: cfg}, nil
}

// Compress compresses src and returns the compressed bytes.
// If the algorithm is None, src is returned unchanged.
func (c *Compressor) Compress(src []byte) ([]byte, error) {
	switch c.cfg.Algorithm {
	case None:
		return src, nil
	case Gzip:
		return c.gzip(src)
	}
	return nil, fmt.Errorf("compressor: unsupported algorithm %q", c.cfg.Algorithm)
}

// Algorithm returns the configured algorithm.
func (c *Compressor) Algorithm() Algorithm {
	return c.cfg.Algorithm
}

func (c *Compressor) gzip(src []byte) ([]byte, error) {
	var buf bytes.Buffer
	level := c.cfg.Level
	if level == 0 {
		level = gzip.DefaultCompression
	}
	w, err := gzip.NewWriterLevel(&buf, level)
\t	return nil, fmt.Errorf("compressor: create gzip writer: %w", err)
	}
	if _, err := io.Copy(w, bytes.NewReader(src)); err != nil {
		return nil, fmt.Errorf("compressor: gzip write: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("compressor: gzip close: %w", err)
	}
	return buf.Bytes(), nil
}
