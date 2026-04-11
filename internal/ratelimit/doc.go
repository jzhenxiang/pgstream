// Package ratelimit provides a token-bucket based rate limiter for
// controlling the throughput of WAL events through the pipeline.
//
// The limiter supports a configurable events-per-second rate and can be
// disabled entirely by setting the rate to zero. It is safe for concurrent
// use and integrates with context cancellation for graceful shutdown.
//
// Example usage:
//
//	limiter, err := ratelimit.New(ratelimit.Config{Rate: 500})
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer limiter.Stop()
//
//	if err := limiter.Wait(ctx); err != nil {
//		return err
//	}
package ratelimit
