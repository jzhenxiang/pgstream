// Package tee implements a best-effort fan-out sink that delivers WAL events
// to every registered downstream sink, collecting and joining errors rather
// than short-circuiting on the first failure.
//
// Use Tee when you want guaranteed delivery attempts to all sinks even if one
// or more are temporarily unavailable, and you intend to handle the combined
// error at a higher level (e.g. via the retry or circuit-breaker packages).
//
// Example:
//
//	t, err := tee.New(kafkaSink, webhookSink)
//	if err != nil {
//		log.Fatal(err)
//	}
//	if err := t.Send(ctx, event); err != nil {
//		// one or more sinks failed
//	}
package tee
