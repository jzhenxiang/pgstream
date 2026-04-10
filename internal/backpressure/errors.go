package backpressure

import "errors"

// ErrBackpressure is returned when Acquire times out waiting for a free slot.
var ErrBackpressure = errors.New("backpressure: acquire timeout exceeded, downstream sink is too slow")
