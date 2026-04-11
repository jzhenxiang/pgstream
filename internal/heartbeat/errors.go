package heartbeat

import "errors"

var errNilSender = errors.New("heartbeat: sender must not be nil")
