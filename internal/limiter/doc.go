// Package limiter provides a token-bucket event rate limiter for use in the
// pgstream pipeline.
//
// It is designed to sit in front of a sink and prevent downstream systems from
// being overwhelmed during high-volume WAL bursts. When limiting is disabled
// (MaxEventsPerSecond == 0) all calls to Allow are no-ops so there is no
// overhead in the common case.
//
// Example usage:
//
//	lim, err := limiter.New(limiter.Config{MaxEventsPerSecond: 1000})
//	if err != nil { ... }
//	if err := lim.Allow(ctx); err != nil {
//		// back-off or drop the event
//	}
package limiter
