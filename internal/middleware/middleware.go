// Package middleware provides HTTP middleware for pgstream's webhook sink,
// including request signing and authentication helpers.
package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

// Signer signs outgoing HTTP requests with an HMAC-SHA256 signature.
type Signer struct {
	secret []byte
	header string
}

// SignerConfig holds configuration for the request signer.
type SignerConfig struct {
	// Secret is the HMAC signing secret.
	Secret string
	// Header is the HTTP header name to write the signature into.
	// Defaults to "X-PGStream-Signature".
	Header string
}

// NewSigner creates a new Signer from the provided config.
func NewSigner(cfg SignerConfig) (*Signer, error) {
	if cfg.Secret == "" {
		return nil, fmt.Errorf("middleware: signing secret must not be empty")
	}
	header := cfg.Header
	if header == "" {
		header = "X-PGStream-Signature"
	}
	return &Signer{
		secret: []byte(cfg.Secret),
		header: header,
	}, nil
}

// Sign computes an HMAC-SHA256 signature over body and sets it on the request.
// The signature format is: sha256=<hex>, prefixed with a Unix timestamp to
// prevent replay attacks: t=<ts>,sha256=<hex>.
func (s *Signer) Sign(req *http.Request, body []byte) {
	ts := fmt.Sprintf("%d", time.Now().Unix())
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(ts))
	mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))
	req.Header.Set(s.header, fmt.Sprintf("t=%s,sha256=%s", ts, sig))
}

// Verify checks whether a signature header value is valid for the given body
// and timestamp tolerance. Returns an error if invalid.
func (s *Signer) Verify(sigHeader string, body []byte, tolerance time.Duration) error {
	var ts, sig string
	_, err := fmt.Sscanf(sigHeader, "t=%s", &ts)
	if err != nil {
		return fmt.Errorf("middleware: malformed signature header")
	}
	// parse manually to handle the comma separator
	for _, part := range splitParts(sigHeader) {
		if len(part) > 2 && part[:2] == "t=" {
			ts = part[2:]
		}
		if len(part) > 7 && part[:7] == "sha256=" {
			sig = part[7:]
		}
	}
	if ts == "" || sig == "" {
		return fmt.Errorf("middleware: missing timestamp or signature")
	}
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(ts))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return fmt.Errorf("middleware: signature mismatch")
	}
	return nil
}

func splitParts(s string) []string {
	var parts []string
	start := 0
	for i, c := range s {
		if c == ',' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}
