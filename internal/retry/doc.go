// Package retry implements a simple exponential-backoff retry helper.
//
// Usage:
//
//	cfg := retry.DefaultConfig()
//	cfg.MaxAttempts = 3
//
//	err := retry.Do(ctx, cfg, func() error {
//		return sink.Send(event)
//	})
//	if err != nil {
//		// all attempts failed
//	}
//
// The delay between attempts grows exponentially from InitialDelay up to
// MaxDelay. Context cancellation aborts the loop immediately.
package retry
