// Package tracer provides lightweight span-based tracing for pgstream pipeline stages.
// Each span records the stage name, start time, duration, and any associated error.
package tracer

import (
	"sync"
	"time"
)

// Span represents a single traced operation.
type Span struct {
	Stage    string
	Start    time.Time
	Duration time.Duration
	Err      error
}

// Tracer collects spans produced during pipeline execution.
type Tracer struct {
	mu    sync.Mutex
	spans []Span
}

// New returns an initialised Tracer.
func New() *Tracer {
	return &Tracer{}
}

// StartSpan begins timing a stage and returns a function that, when called,
// records the completed span. Pass a non-nil error to mark the span as failed.
//
//	finish := t.StartSpan("decode")
//	defer finish(nil)
func (t *Tracer) StartSpan(stage string) func(err error) {
	start := time.Now()
	return func(err error) {
		span := Span{
			Stage:    stage,
			Start:    start,
			Duration: time.Since(start),
			Err:      err,
		}
		t.mu.Lock()
		t.spans = append(t.spans, span)
		t.mu.Unlock()
	}
}

// Spans returns a copy of all recorded spans.
func (t *Tracer) Spans() []Span {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]Span, len(t.spans))
	copy(out, t.spans)
	return out
}

// Reset clears all recorded spans.
func (t *Tracer) Reset() {
	t.mu.Lock()
	t.spans = t.spans[:0]
	t.mu.Unlock()
}

// ErrorCount returns the number of spans that recorded a non-nil error.
func (t *Tracer) ErrorCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	count := 0
	for _, s := range t.spans {
		if s.Err != nil {
			count++
		}
	}
	return count
}
