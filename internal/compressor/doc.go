// Package compressor provides transparent payload compression for pgstream
// event data before it is forwarded to sinks.
//
// Supported algorithms:
//
//   - none  – passthrough, no compression applied (default)
//   - gzip  – standard DEFLATE-based compression
//
// Usage:
//
//	c, err := compressor.New(compressor.Config{Algorithm: compressor.Gzip})
//	if err != nil { ... }
//	compressed, err := c.Compress(payload)
//
// The compressor is stateless and safe for concurrent use.
package compressor
