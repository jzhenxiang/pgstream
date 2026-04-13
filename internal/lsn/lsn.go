// Package lsn provides utilities for parsing, formatting, and comparing
// PostgreSQL Log Sequence Numbers (LSNs).
package lsn

import (
	"fmt"
	"strconv"
	"strings"
)

// LSN represents a PostgreSQL Log Sequence Number.
type LSN uint64

// Zero is the zero value for an LSN.
const Zero LSN = 0

// String returns the standard PostgreSQL LSN representation (e.g. "0/1A2B3C4D").
func (l LSN) String() string {
	return fmt.Sprintf("%X/%X", uint32(l>>32), uint32(l))
}

// Parse parses a PostgreSQL LSN string (e.g. "0/1A2B3C4D") into an LSN.
func Parse(s string) (LSN, error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return Zero, fmt.Errorf("lsn: invalid format %q: expected high/low", s)
	}

	high, err := strconv.ParseUint(parts[0], 16, 32)
	if err != nil {
		return Zero, fmt.Errorf("lsn: invalid high segment %q: %w", parts[0], err)
	}

	low, err := strconv.ParseUint(parts[1], 16, 32)
	if err != nil {
		return Zero, fmt.Errorf("lsn: invalid low segment %q: %w", parts[1], err)
	}

	return LSN((high << 32) | low), nil
}

// MustParse parses an LSN string and panics on error. Intended for tests and
// compile-time constants.
func MustParse(s string) LSN {
	l, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return l
}

// After reports whether l is strictly greater than other.
func (l LSN) After(other LSN) bool {
	return l > other
}

// Before reports whether l is strictly less than other.
func (l LSN) Before(other LSN) bool {
	return l < other
}

// IsZero reports whether the LSN is the zero value.
func (l LSN) IsZero() bool {
	return l == Zero
}

// Max returns the larger of l and other.
func Max(a, b LSN) LSN {
	if a > b {
		return a
	}
	return b
}
