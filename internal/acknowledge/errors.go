package acknowledge

import "errors"

// ErrNilSender is returned when a nil Sender is provided to New.
var ErrNilSender = errors.New("acknowledge: sender must not be nil")
