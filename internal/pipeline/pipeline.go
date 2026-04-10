// Package pipeline wires together the WAL reader, filter, transform, and sink
// into a single cohesive processing pipeline.
package pipeline

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/pgstream/pgstream/internal/filter"
	"github.com/pgstream/pgstream/internal/metrics"
	"github.com/pgstream/pgstream/internal/sink"
	"github.com/pgstream/pgstream/internal/transform"
	"github.com/pgstream/pgstream/internal/wal"
)

// Pipeline orchestrates the flow of WAL events from reader to sink.
type Pipeline struct {
	processor *wal.Processor
	metrics   *metrics.Metrics
	logger    *slog.Logger
}

// Config holds all dependencies needed to build a Pipeline.
type Config struct {
	Reader    *wal.Reader
	Filter    *filter.Filter
	Transform *transform.Transform
	Sink      sink.Sink
	Metrics   *metrics.Metrics
	Logger    *slog.Logger
}

// New constructs a Pipeline from the provided Config.
func New(cfg Config) (*Pipeline, error) {
	if cfg.Reader == nil {
		return nil, fmt.Errorf("pipeline: reader is required")
	}
	if cfg.Sink == nil {
		return nil, fmt.Errorf("pipeline: sink is required")
	}

	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	m := cfg.Metrics
	if m == nil {
		m = metrics.New()
	}

	proc, err := wal.NewProcessor(cfg.Reader, cfg.Filter, cfg.Transform, cfg.Sink, m, logger)
	if err != nil {
		return nil, fmt.Errorf("pipeline: failed to create processor: %w", err)
	}

	return &Pipeline{
		processor: proc,
		metrics:   m,
		logger:    logger,
	}, nil
}

// Run starts the pipeline and blocks until ctx is cancelled or a fatal error
// occurs. It logs the cause when the context is cancelled.
func (p *Pipeline) Run(ctx context.Context) error {
	p.logger.Info("pipeline starting")
	if err := p.processor.Run(ctx); err != nil {
		if ctx.Err() != nil {
			p.logger.Info("pipeline stopped", "reason", ctx.Err())
			return nil
		}
		return fmt.Errorf("pipeline: %w", err)
	}
	p.logger.Info("pipeline stopped")
	return nil
}

// Metrics returns the metrics instance used by this pipeline.
func (p *Pipeline) Metrics() *metrics.Metrics {
	return p.metrics
}
