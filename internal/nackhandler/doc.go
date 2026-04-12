// Package nackhandler implements a negative-acknowledgement handler for the
// pgstream pipeline.
//
// When a sink fails to deliver a WAL event the pipeline can call Handle with
// the offending event and the error. The handler tracks per-event delivery
// attempts and, once the configured maximum is exceeded, forwards the event
// to a fallback sink (typically a dead-letter queue) so that the main pipeline
// can continue without blocking.
//
// Usage:
//
//	h, err := nackhandler.New(nackhandler.Config{
//		MaxAttempts: 3,
//		Fallback:    dlqSink,
//	})
//	if err != nil { ... }
//
//	if sendErr := primary.Send(ctx, event); sendErr != nil {
//		if err := h.Handle(ctx, event, sendErr); err != nil {
//			log.Printf("nack: %v", err)
//		}
//	}
package nackhandler
