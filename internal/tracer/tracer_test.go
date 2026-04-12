package tracer_test

import (
	"errors"
	"testing"
	"time"

	"github.com/your-org/pgstream/internal/tracer"
)

func TestNew_ReturnsEmptyTracer(t *testing.T) {
	tr := tracer.New()
	if spans := tr.Spans(); len(spans) != 0 {
		t.Fatalf("expected 0 spans, got %d", len(spans))
	}
}

func TestStartSpan_RecordsSpan(t *testing.T) {
	tr := tracer.New()
	finish := tr.StartSpan("decode")
	time.Sleep(time.Millisecond)
	finish(nil)

	spans := tr.Spans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Stage != "decode" {
		t.Errorf("expected stage 'decode', got %q", spans[0].Stage)
	}
	if spans[0].Duration <= 0 {
		t.Error("expected positive duration")
	}
	if spans[0].Err != nil {
		t.Errorf("expected nil error, got %v", spans[0].Err)
	}
}

func TestStartSpan_RecordsError(t *testing.T) {
	tr := tracer.New()
	sentinel := errors.New("sink unavailable")
	finish := tr.StartSpan("send")
	finish(sentinel)

	spans := tr.Spans()
	if spans[0].Err != sentinel {
		t.Errorf("expected sentinel error, got %v", spans[0].Err)
	}
}

func TestErrorCount(t *testing.T) {
	tr := tracer.New()
	tr.StartSpan("a")(nil)
	tr.StartSpan("b")(errors.New("boom"))
	tr.StartSpan("c")(errors.New("bang"))

	if got := tr.ErrorCount(); got != 2 {
		t.Errorf("expected 2 errors, got %d", got)
	}
}

func TestReset_ClearsSpans(t *testing.T) {
	tr := tracer.New()
	tr.StartSpan("x")(nil)
	tr.Reset()

	if spans := tr.Spans(); len(spans) != 0 {
		t.Fatalf("expected 0 spans after reset, got %d", len(spans))
	}
}

func TestSpans_ReturnsCopy(t *testing.T) {
	tr := tracer.New()
	tr.StartSpan("y")(nil)

	s1 := tr.Spans()
	s1[0].Stage = "mutated"

	s2 := tr.Spans()
	if s2[0].Stage == "mutated" {
		t.Error("Spans() should return a copy, not a reference")
	}
}
