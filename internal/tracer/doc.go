// Package tracer provides a minimal, allocation-light span collector for
// instrumenting pgstream pipeline stages without introducing an external
// tracing dependency.
//
// Usage:
//
//	t := tracer.New()
//
//	finish := t.StartSpan("wal.decode")
//	// ... do work ...
//	finish(err) // nil on success
//
//	for _, span := range t.Spans() {
//		fmt.Printf("%s took %s\n", span.Stage, span.Duration)
//	}
//
// Tracer is safe for concurrent use.
package tracer
