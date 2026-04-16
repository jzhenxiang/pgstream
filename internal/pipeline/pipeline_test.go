package pipeline_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pgstream/pgstream/internal/pipeline"
	"github.com/pgstream/pgstream/internal/wal"
)

func TestNew_NilReader(t *testing.T) {
	_, err := pipeline.New(pipeline.Config{
		Reader: nil,
		Sink:   &stubSink{},
	})
	if err == nil {
		t.Fatal("expected error for nil reader, got nil")
	}
}

func TestNew_NilSink(t *testing.T) {
	_, err := pipeline.New(pipeline.Config{
		Reader: &wal.Reader{},
		Sink:   nil,
	})
	if err == nil {
		t.Fatal("expected error for nil sink, got nil")
	}
}

func TestNew_Valid(t *testing.T) {
	p, err := pipeline.New(pipeline.Config{
		Reader: &wal.Reader{},
		Sink:   &stubSink{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil pipeline")
	}
}

func TestPipeline_Metrics_NotNil(t *testing.T) {
	p, err := pipeline.New(pipeline.Config{
		Reader: &wal.Reader{},
		Sink:   &stubSink{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Metrics() == nil {
		t.Fatal("expected non-nil metrics")
	}
}

func TestPipeline_Run_ContextCancelled(t *testing.T) {
	p, err := pipeline.New(pipeline.Config{
		Reader: &wal.Reader{},
		Sink:   &stubSink{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := p.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestNew_NilReaderAndSink verifies that an error is returned when both reader
// and sink are nil.
func TestNew_NilReaderAndSink(t *testing.T) {
	_, err := pipeline.New(pipeline.Config{
		Reader: nil,
		Sink:   nil,
	})
	if err == nil {
		t.Fatal("expected error for nil reader and sink, got nil")
	}
}

// stubSink satisfies sink.Sink for testing.
type stubSink struct{}

func (s *stubSink) Send(_ context.Context, _ []byte) error { return nil }
