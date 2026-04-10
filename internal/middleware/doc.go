// Package middleware provides reusable HTTP middleware components for pgstream.
//
// It includes:
//
//   - Signer: signs outgoing webhook HTTP requests with HMAC-SHA256 so that
//     receivers can verify the payload originated from pgstream. The signature
//     is placed in a configurable header (default: X-PGStream-Signature) using
//     the format "t=<unix_ts>,sha256=<hex>".
//
//   - Logger: a request-logging middleware that emits structured slog entries
//     for every HTTP request, including method, path, status code and latency.
//
//   - Recovery: a panic-recovery middleware that catches unexpected panics and
//     returns a 500 Internal Server Error instead of crashing the process.
//
// Usage example:
//
//	mux := http.NewServeMux()
//	mux.HandleFunc("/health", healthHandler)
//	handler := middleware.Recovery(logger)(middleware.Logger(logger)(mux))
//	http.ListenAndServe(":8080", handler)
package middleware
