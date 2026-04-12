package window

import "errors"

// ErrInvalidSize is returned when the window size is not positive.
var ErrInvalidSize = errors.New("window: size must be greater than zero")

// ErrInvalidGranule is returned when the granularity is invalid.
var ErrInvalidGranule = errors.New("window: granule must be > 0 and <= size")
